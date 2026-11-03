package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys/lystime"
)

type background struct {
	wg sync.WaitGroup
}

// loadState loads persisted blocked IPs from db, if any.
func (srvApp *httpServerApplication) loadState(ctx context.Context) error {
	if err := srvApp.loadBlockedIPsFromDb(ctx); err != nil {
		return fmt.Errorf("initialization: failed to load blocked IPs from db: %w", err)
	}
	return nil
}

// startBackground starts the periodic maintenance goroutines. They stop when ctx is canceled.
func (srvApp *httpServerApplication) startBackground(ctx context.Context) *background {
	bg := &background{}

	bg.wg.Go(func() {
		lystime.RunEvery(ctx, 30*time.Minute, func() { srvApp.RateLimits.CleanupExpired() })
	})

	return bg
}

// serve loads state, starts background goroutines and serves on ln until ctx is canceled or the server fails.
// It then shuts everything down and returns the collected errors. stop must cancel ctx.
func (srvApp *httpServerApplication) serve(ctx context.Context, stop context.CancelFunc, srv *http.Server, ln net.Listener) error {
	defer ln.Close()

	if err := srvApp.loadState(ctx); err != nil {
		return err
	}

	bg := srvApp.startBackground(ctx)

	srvErrCh := make(chan error, 1)
	go func() {
		cmd.StartupMsg("suppsrv", srvApp.Config.Api.Port, srvApp.Config.General.Debug, srvApp.Logger)
		srvErrCh <- srv.Serve(ln)
	}()

	var serveErr error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = fmt.Errorf("srv.Serve failed: %w", err)
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	stop()

	return errors.Join(serveErr, srvApp.shutdown(srv, bg))
}

// shutdown stops everything started by serve. The context passed to startBackground must be canceled first.
func (srvApp *httpServerApplication) shutdown(srv *http.Server, bg *background) error {
	bg.wg.Wait()
	return cmd.ShutdownHttp(srv, 10*time.Second)
}
