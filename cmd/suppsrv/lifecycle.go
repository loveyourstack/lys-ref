package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type background struct {
	wg sync.WaitGroup
}

// newHttpServer wraps handler in a server with timeouts.
func newHttpServer(handler http.Handler, addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           http.TimeoutHandler(handler, 5*time.Second, "request timed out"),
		IdleTimeout:       time.Second,
		MaxHeaderBytes:    1024 * 1024, // 1 MB
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
		WriteTimeout:      10 * time.Second,
	}
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
	bg.wg.Go(func() { srvApp.runRateLimitCleanup(ctx, 30*time.Minute) })
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

	startupMsg := fmt.Sprintf("starting suppsrv on port: %s", srvApp.Config.Api.Port)
	if srvApp.Config.General.Debug {
		startupMsg += ", debug: true"
	}
	srvApp.Logger.Info(startupMsg)

	srvErrCh := make(chan error, 1)
	go func() {
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
	return srvApp.shutdownHttp(srv)
}

// shutdownHttp gracefully shuts down the server, waiting for active requests to finish or timeout before forcefully closing.
func (srvApp *httpServerApplication) shutdownHttp(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	if closeErr := srv.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("srv.Close failed: %w", closeErr))
	}

	return fmt.Errorf("shutdown: srv.Shutdown failed: %w", err)
}
