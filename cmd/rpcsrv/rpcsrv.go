package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssvcaccount"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	configFilePath := flag.String("configFilePath", "ref_config.toml", "Path to the config file")
	flag.Parse()

	if err := run(*configFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(configFilePath string) error {

	// load config from file
	conf := myapp.Config{}
	err := conf.LoadFromFile(configFilePath)
	if err != nil {
		return fmt.Errorf("conf.LoadFromFile (%s) failed: %w", configFilePath, err)
	}

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create gRPC server app
	rpcApp := &gRpcServerApplication{
		Application: app,

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3
	}

	// attach Get options
	rpcApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		FieldsParamName:  "fields",
		PageParamName:    "page",
		PerPageParamName: "per_page",
		SortParamName:    "sorts",
	})
	if err != nil {
		return fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	// attach auth
	if rpcApp.Config.General.Env == appenv.Prod {
		rpcApp.UseXForwardedFor = true
		rpcApp.XForwardedForIdx = 0
	}
	rpcApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                                                     // block after 5 failed attempts
	rpcApp.Sessions = lysauth.NewAppSessions(rpcApp.Validate, 10*time.Hour, rpcApp.UseXForwardedFor, rpcApp.XForwardedForIdx) // allow 10 hour sessions

	// connect to db and assign pool to rpcApp
	rpcApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbRpcUser, conf.General.AppName+" Rpc Srv", myapp.TypesToRegister)
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPoolWithTypes failed: %w", err)
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
	rpcApp.SvcAccountStore = syssvcaccount.Store{Db: rpcApp.Db}
	rpcApp.UserStore = sysuser.Store{Db: rpcApp.Db}

	// create TCP listener for the gRPC server (loopback: in prod, is behind nginx)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", conf.RpcApi.Port))
	if err != nil {
		return fmt.Errorf("net.Listen failed: %w", err)
	}
	defer lis.Close()

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
		return fmt.Errorf("rpcApp.loadSessionsFromDb failed: %w", err)
	}
	err = rpcApp.loadLoginAttemptsFromDb(ctx)
	if err != nil {
		return fmt.Errorf("rpcApp.loadLoginAttemptsFromDb failed: %w", err)
	}

	// create wait group for background goroutines that must exit before sessions are persisted
	var bgWg sync.WaitGroup

	// archive expired sessions immediately, and start goroutine to periodically archive them
	rpcApp.archiveExpiredSessions(ctx)
	bgWg.Go(func() { rpcApp.runExpiredSessionArchiver(ctx, 5*time.Minute) })

	// start goroutine to periodically clean up expired rate limit buckets
	bgWg.Go(func() { rpcApp.runRateLimitCleanup(ctx, 30*time.Minute) })

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
	var serveErr error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr = fmt.Errorf("grpcServer.Serve failed: %w", err)
		}
	case <-ctx.Done():
		rpcApp.Logger.Info("shutdown signal received")
	}

	// stop background goroutines, restore default signal handling
	stop()

	// wait for background goroutines to exit so none are mid-archive when sessions are persisted
	bgWg.Wait()

	// every shutdown step runs regardless of earlier failures; errors are collected and returned at the end
	runErr := serveErr

	// --------------------------------

	stopped := make(chan struct{})
	go func() {
		// gracefully shutdown server, waiting for active requests to finish
		grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		grpcServer.Stop()
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: graceful stop timed out, forcing stop"))
	}

	persistCtx, persistCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer persistCancel()

	// persist sessions and login attempts to db, for loading on restart
	if err := rpcApp.persistSessions(persistCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: rpcApp.persistSessions failed: %w", err))
	}
	if err := rpcApp.persistLoginAttempts(persistCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: rpcApp.persistLoginAttempts failed: %w", err))
	}

	return runErr
}

func toPbMetadata(m lysapi.SelectMetadata, itemCount int) *pb.GetMetadata {
	return &pb.GetMetadata{
		Count:                 int32(itemCount),
		TotalCount:            m.TotalCount,
		TotalCountIsEstimated: m.TotalCountIsEstimated,
	}
}

// pbUnpagedMetadata returns metadata for a list that was returned in full, without paging.
func pbUnpagedMetadata(n int) *pb.GetMetadata {
	return &pb.GetMetadata{Count: int32(n), TotalCount: int64(n)}
}
