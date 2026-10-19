package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
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

	ctx := context.Background()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create rRPC server app
	rpcSrvApp := &gRpcServerApplication{
		Application: app,
	}

	// attach Get options
	rpcSrvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		FieldsParamName:  "fields",
		PageParamName:    "page",
		PerPageParamName: "per_page",
		SortParamName:    "sorts",
	})
	if err != nil {
		log.Fatalf("initialization: failed to fill GetOptions: %s", err.Error())
	}

	// attach auth
	if rpcSrvApp.Config.General.Env == appenv.Prod {
		rpcSrvApp.UseXForwardedFor = true
		rpcSrvApp.XForwardedForIdx = 0
	}
	rpcSrvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                                                              // block after 5 failed attempts
	rpcSrvApp.Sessions = lysauth.NewAppSessions(rpcSrvApp.Validate, 10*time.Hour, rpcSrvApp.UseXForwardedFor, rpcSrvApp.XForwardedForIdx) // allow 10 hour sessions

	// connect to db and assign pool to rpcSrvApp
	rpcSrvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbRpcUser, rpcSrvApp.Config.General.AppName+" Rpc Srv", myapp.TypesToRegister)
	if err != nil {
		log.Fatalf("initialization: failed to create regular db connection pool: %s", err.Error())
	}
	defer rpcSrvApp.Db.Close()

	// attach stores
	rpcSrvApp.BlockedIPStore = sysblockedip.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.CountryStore = geocountry.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.GeoLocationStore = mmlocation.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.GeoNetworkStore = mmnetwork.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.LoginAttemptStore = sysrpcloginattempt.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.SessionStore = sysrpcsession.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.SessionHistStore = sysrpcsessionhist.Store{Db: rpcSrvApp.Db}
	//rpcSrvApp.SrvLogStore = syssrvreq.Store{Db: rpcSrvApp.Db}
	rpcSrvApp.UserStore = sysuser.Store{Db: rpcSrvApp.Db}

	// create TCP listener for the gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", conf.RpcApi.Port))
	if err != nil {
		log.Fatalf("initialization: net.Listen failed: %v", err)
	}

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting refrpcsrv on port: %s", conf.RpcApi.Port)
	if conf.General.Debug {
		startupMsg += ", debug: true"
	}
	rpcSrvApp.Logger.Info(startupMsg)

	// create gRPC server
	var opts []grpc.ServerOption

	// use interceptor to distinguish between expected client errors and unexpected server errors
	opts = append(opts, grpc.UnaryInterceptor(rpcSrvApp.unaryErrorInterceptor))

	grpcServer := grpc.NewServer(opts...)

	// register services
	pb.RegisterAuthSvcServer(grpcServer, &authSrv{gRpcServerApplication: rpcSrvApp})

	campStore := dmcampaign.Store{Db: rpcSrvApp.Db}
	pb.RegisterDmCampaignSvcServer(grpcServer, &dmCampSrv{gRpcServerApplication: rpcSrvApp, store: campStore,
		jsonKeyDbNameMap: campStore.GetPlan().JsonKeyDbNameMap()})

	vertStore := dmvertical.Store{Db: rpcSrvApp.Db}
	pb.RegisterDmVerticalSvcServer(grpcServer, &dmVertServer{gRpcServerApplication: rpcSrvApp, store: vertStore,
		jsonKeyDbNameMap: vertStore.GetPlan().JsonKeyDbNameMap()})

	// if debug mode is enabled, enable server reflection for test clients
	if conf.General.Debug {
		reflection.Register(grpcServer)
	}

	// start gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func toPbMetadata(m lysapi.SelectMetadata, itemCount int) *pb.GetMetadata {
	return &pb.GetMetadata{
		Count:                 int32(itemCount),
		TotalCount:            m.TotalCount,
		TotalCountIsEstimated: m.TotalCountIsEstimated,
	}
}
