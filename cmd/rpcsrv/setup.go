package main

import (
	"context"
	"fmt"
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
	"github.com/loveyourstack/lys/lysrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// newBaseApplication builds the in-memory part of the application.
func newBaseApplication(conf *myapp.Config) (rpcApp *gRpcServerApplication, err error) {

	rpcApp = &gRpcServerApplication{
		Application: cmd.NewApplication(conf),

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3

		UseRealIpHeader: conf.General.Env == appenv.Prod,

		ShutdownTimeouts: defaultShutdownTimeouts,
	}

	rpcApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		FieldsParamName:  "fields",
		PageParamName:    "page",
		PerPageParamName: "per_page",
		SortParamName:    "sorts",
	})
	if err != nil {
		return nil, fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	rpcApp.attachAuth()

	return rpcApp, nil
}

// newGrpcServerApplication builds the application and connects everything it depends on.
// On error, anything already opened is closed. On success, the caller must call Close.
func newGrpcServerApplication(ctx context.Context, conf *myapp.Config) (rpcApp *gRpcServerApplication, err error) {

	// build the in-memory part of the application first
	rpcApp, err = newBaseApplication(conf)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			rpcApp.Close()
			rpcApp = nil
		}
	}()

	rpcApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbRpcUser, conf.General.AppName+" Rpc Srv", myapp.TypesToRegister)
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPoolWithTypes failed: %w", err)
	}

	rpcApp.attachStores()

	return rpcApp, nil
}

// Close releases the db pool. It is safe to call on a partially built app.
func (rpcApp *gRpcServerApplication) Close() {
	if rpcApp.Db != nil {
		rpcApp.Db.Close()
	}
}

// attachAuth must be called after UseRealIpHeader is set.
func (rpcApp *gRpcServerApplication) attachAuth() {
	rpcApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                           // block after 5 failed attempts
	rpcApp.Sessions = lysauth.NewAppSessions(rpcApp.Validate, 10*time.Hour, rpcApp.UseRealIpHeader) // allow 10 hour sessions
}

func (rpcApp *gRpcServerApplication) attachStores() {
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
}

func newGrpcServer(rpcApp *gRpcServerApplication) *grpc.Server {
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		lysrpc.SetClientIp(rpcApp.UseRealIpHeader, rpcApp.Logger), // add client IP to ctx, for use later in the chain
		rpcApp.logRequest,              // initialize userInfo in context and record final status and duration
		lysrpc.MapError(rpcApp.Logger), // convert errors to gRPC status codes
		lysrpc.RejectBlockedIp(rpcApp.LoginAttempts, rpcApp.Logger), // check if the request comes from a blocked IP
		rpcApp.limitUnauthed, // rate limit public methods by IP
		rpcApp.authenticate,  // authenticate requests and update userInfo in context
		rpcApp.limitAuthed,   // rate limit authenticated methods by user ID
	))

	pb.RegisterAuthSvcServer(grpcServer, &authSrv{gRpcServerApplication: rpcApp,
		laStore: rpcApp.LoginAttemptStore, sessStore: rpcApp.SessionStore})

	campStore := dmcampaign.Store{Db: rpcApp.Db}
	pb.RegisterDmCampaignSvcServer(grpcServer, &dmCampSrv{gRpcServerApplication: rpcApp, store: campStore,
		jsonKeyDbNameMap: campStore.GetPlan().JsonKeyDbNameMap()})

	vertStore := dmvertical.Store{Db: rpcApp.Db}
	pb.RegisterDmVerticalSvcServer(grpcServer, &dmVertServer{gRpcServerApplication: rpcApp, store: vertStore,
		jsonKeyDbNameMap: vertStore.GetPlan().JsonKeyDbNameMap()})

	if rpcApp.Config.General.Debug {
		reflection.Register(grpcServer)
	}

	return grpcServer
}
