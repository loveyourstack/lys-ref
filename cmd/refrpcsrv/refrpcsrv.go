package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/cmd/refrpcsrv/pb"
	"github.com/loveyourstack/lys-ref/internal/myapp"
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

	// connect to db and assign pool to rpcSrvApp
	rpcSrvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbServerUser, rpcSrvApp.Config.General.AppName+" Srv", myapp.TypesToRegister)
	if err != nil {
		log.Fatalf("initialization: failed to create regular db connection pool: %s", err.Error())
	}
	defer rpcSrvApp.Db.Close()

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
	grpcServer := grpc.NewServer(opts...)

	// register services
	pb.RegisterCampaignServiceServer(grpcServer, &campaignServer{gRpcServerApplication: rpcSrvApp})
	pb.RegisterVerticalServiceServer(grpcServer, &verticalServer{gRpcServerApplication: rpcSrvApp})

	// if debug mode is enabled, enable server reflection
	if conf.General.Debug {
		reflection.Register(grpcServer)
	}

	// start gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
