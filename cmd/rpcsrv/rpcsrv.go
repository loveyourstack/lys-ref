package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/pb"
	"github.com/loveyourstack/lys/lysapi"
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
	if err := conf.LoadFromFile(configFilePath); err != nil {
		return fmt.Errorf("conf.LoadFromFile (%s) failed: %w", configFilePath, err)
	}

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// build and wire up the server app
	rpcApp, err := newGrpcServerApplication(ctx, &conf)
	if err != nil {
		return err
	}
	defer rpcApp.Close()

	// create TCP listener for the gRPC server (loopback: in prod, is behind nginx)
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", conf.RpcApi.Port))
	if err != nil {
		return fmt.Errorf("net.Listen failed: %w", err)
	}

	// start serving gRPC requests
	return rpcApp.serve(ctx, stop, newGrpcServer(rpcApp), lis)
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
