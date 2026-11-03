package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys/lysauth"
)

func main() {
	configFilePath := flag.String("configFilePath", "ref_config.toml", "Path to the config file")
	flag.Parse()

	// use run func pattern: log.Fatal/os.Exit ignores deferred calls
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

	// build and wire up the server app
	srvApp, err := newHttpServerApplication(ctx, &conf)
	if err != nil {
		return err
	}
	defer srvApp.Close()

	//srv := newHttpServer(srvApp.getRouter(), "127.0.0.1:"+conf.Api.Port)

	// create HTTP server with timeout bypass for WebSocket requests
	srv := cmd.NewHttpServer(srvApp.getRouter(), "127.0.0.1:"+conf.Api.Port, func(r *http.Request) bool {
		return lysauth.IsWebSocket(r.Header)
	})

	// listen before starting anything else so that an unavailable port fails fast
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("net.Listen failed: %w", err)
	}

	// start serving HTTP requests
	return srvApp.serve(ctx, stop, srv, ln)
}
