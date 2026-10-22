package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/cmd/rpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	configFileName := "ref_config.toml"

	// mandatory flag if not using default
	configFilePath := flag.String("configFilePath", configFileName, "Path to the config file")

	flag.Parse()

	// load config from file
	conf := myapp.Config{}
	err := conf.LoadFromFile(*configFilePath)
	if err != nil {
		log.Fatalf("initialization: conf.LoadFromFile (%s) failed: %s", *configFilePath, err.Error())
	}

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create rRPC server app
	rpcApp := &gRpcServerApplication{
		Application: app,

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3
	}

	// start goroutine to periodically clean up expired rate limit buckets
	go rpcApp.runRateLimitCleanup(ctx, 30*time.Minute)

	// attach Get options
	rpcApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		FieldsParamName:  "fields",
		PageParamName:    "page",
		PerPageParamName: "per_page",
		SortParamName:    "sorts",
	})
	if err != nil {
		log.Fatalf("initialization: failed to fill GetOptions: %s", err.Error())
	}

	// attach auth
	if rpcApp.Config.General.Env == appenv.Prod {
		rpcApp.UseXForwardedFor = true
		rpcApp.XForwardedForIdx = 0
	}
	rpcApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                                                     // block after 5 failed attempts
	rpcApp.Sessions = lysauth.NewAppSessions(rpcApp.Validate, 10*time.Hour, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx) // allow 10 hour sessions

	// connect to db and assign pool to rpcSrvApp
	rpcApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbRpcUser, rpcApp.Config.General.AppName+" Rpc Srv", myapp.TypesToRegister)
	if err != nil {
		log.Fatalf("initialization: failed to create regular db connection pool: %s", err.Error())
	}
	defer rpcApp.Db.Close()

	// attach stores
	rpcApp.BlockedIPStore = sysblockedip.Store{Db: rpcApp.Db}
	rpcApp.CountryStore = geocountry.Store{Db: rpcApp.Db}
	rpcApp.GeoLocationStore = mmlocation.Store{Db: rpcApp.Db}
	rpcApp.GeoNetworkStore = mmnetwork.Store{Db: rpcApp.Db}
	rpcApp.LoginAttemptStore = sysrpcloginattempt.Store{Db: rpcApp.Db}
	rpcApp.SessionStore = sysrpcsession.Store{Db: rpcApp.Db}
	rpcApp.SessionHistStore = sysrpcsessionhist.Store{Db: rpcApp.Db}
	rpcApp.SrvLogStore = sysrpcreq.Store{Db: rpcApp.Db}
	rpcApp.UserStore = sysuser.Store{Db: rpcApp.Db}

	// create TCP listener for the gRPC server (loopback: in prod, is behind nginx)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", conf.RpcApi.Port))
	if err != nil {
		log.Fatalf("initialization: net.Listen failed: %v", err)
	}

	// create gRPC server
	var opts []grpc.ServerOption

	// define interceptor chain
	opts = append(opts, grpc.ChainUnaryInterceptor(
		rpcApp.logRequest,      // initialize userInfo in context and record final status and duration
		rpcApp.mapError,        // convert errors to gRPC status codes
		rpcApp.rejectBlockedIp, // check if the request comes from a blocked IP
		rpcApp.limitUnauthed,   // rate limit public methods by IP
		rpcApp.authenticate,    // authenticate requests and update userInfo in context
		rpcApp.limitAuthed,     // rate limit authenticated methods by user ID
	))

	grpcServer := grpc.NewServer(opts...)

	// register services
	pb.RegisterAuthSvcServer(grpcServer, &authSrv{gRpcServerApplication: rpcApp,
		laStore: rpcApp.LoginAttemptStore, sessStore: rpcApp.SessionStore})

	campStore := dmcampaign.Store{Db: rpcApp.Db}
	pb.RegisterDmCampaignSvcServer(grpcServer, &dmCampSrv{gRpcServerApplication: rpcApp, store: campStore,
		jsonKeyDbNameMap: campStore.GetPlan().JsonKeyDbNameMap()})

	vertStore := dmvertical.Store{Db: rpcApp.Db}
	pb.RegisterDmVerticalSvcServer(grpcServer, &dmVertServer{gRpcServerApplication: rpcApp, store: vertStore,
		jsonKeyDbNameMap: vertStore.GetPlan().JsonKeyDbNameMap()})

	// if debug mode is enabled, enable server reflection for test clients
	if conf.General.Debug {
		reflection.Register(grpcServer)
	}

	// load persisted sessions and login attempts from db, if any
	err = rpcApp.loadSessionsFromDb(ctx)
	if err != nil {
		log.Fatalf("initialization: rpcApp.loadSessionsFromDb failed: %s", err.Error())
	}
	err = rpcApp.loadLoginAttemptsFromDb(ctx)
	if err != nil {
		log.Fatalf("initialization: rpcApp.loadLoginAttemptsFromDb failed: %s", err.Error())
	}

	// archive expired sessions immediately, and start goroutine to periodically archive them
	rpcApp.archiveExpiredSessions(ctx)
	go rpcApp.runExpiredSessionArchiver(ctx, 5*time.Minute)

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting rpcsrv on port: %s", conf.RpcApi.Port)
	if conf.General.Debug {
		startupMsg += ", debug: true"
	}
	rpcApp.Logger.Info(startupMsg)

	// start server in a goroutine so that it doesn't block the main thread, which is waiting for shutdown signals
	srvErrCh := make(chan error, 1)
	go func() {
		srvErrCh <- grpcServer.Serve(lis)
	}()

	// wait for shutdown signal or server error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			rpcApp.Logger.Error("grpcServer.Serve failed", "error", err)
		}
	case <-ctx.Done():
		rpcApp.Logger.Info("shutdown signal received")
	}

	// --------------------------------

	// create context with timeout for server shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		// gracefully shutdown server, waiting for active requests to finish
		grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		rpcApp.Logger.Error("shutdown: graceful stop timed out, forcing stop")
		grpcServer.Stop()
	}

	// persist sessions and login attempts to db, for loading on restart
	if err := rpcApp.persistSessions(context.Background()); err != nil {
		rpcApp.Logger.Error("shutdown: rpcApp.persistSessions failed", "error", err)
	}
	if err := rpcApp.persistLoginAttempts(context.Background()); err != nil {
		rpcApp.Logger.Error("shutdown: rpcApp.persistLoginAttempts failed", "error", err)
	}
}

func toPbMetadata(m lysapi.SelectMetadata, itemCount int) *pb.GetMetadata {
	return &pb.GetMetadata{
		Count:                 int32(itemCount),
		TotalCount:            m.TotalCount,
		TotalCountIsEstimated: m.TotalCountIsEstimated,
	}
}
