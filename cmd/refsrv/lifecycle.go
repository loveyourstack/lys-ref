package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lysws"
)

// maxUserWsConnections is the maximum number of WebSocket connections allowed per user
const maxUserWsConnections = 5

// shutdownTimeouts are the maximum durations of the individual shutdown steps
type shutdownTimeouts struct {
	Http    time.Duration // graceful http shutdown, after which remaining connections are closed
	Persist time.Duration // persisting sessions and login attempts
	Hub     time.Duration // waiting for the notification hub listener to exit
}

var defaultShutdownTimeouts = shutdownTimeouts{
	Http:    10 * time.Second,
	Persist: 10 * time.Second,
	Hub:     5 * time.Second,
}

// background holds the goroutines started by startBackground
type background struct {
	wg      sync.WaitGroup // goroutines that must exit before sessions are persisted
	hubDone chan struct{}  // closed when the notification hub listener exits
}

// newHttpServer wraps handler in a server with timeouts
func newHttpServer(handler http.Handler, addr string) *http.Server {

	timedHandler := http.TimeoutHandler(handler, 5*time.Second, "request timed out")

	return &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// http.TimeoutHandler does not implement the http.Hijacker interface, which is needed for WebSocket upgrades
			// so allow websocket requests to skip the timeout handler
			if lysauth.IsWebSocket(r.Header) {
				handler.ServeHTTP(w, r)
				return
			}

			// http requests
			timedHandler.ServeHTTP(w, r)
		}),
		IdleTimeout:       time.Second,
		MaxHeaderBytes:    1024 * 1024, // 1 MB
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
		WriteTimeout:      10 * time.Second,
	}
}

// loadState loads persisted sessions and login attempts from db, if any
func (srvApp *httpServerApplication) loadState(ctx context.Context) error {

	if err := srvApp.loadSessionsFromDb(ctx); err != nil {
		return fmt.Errorf("srvApp.loadSessionsFromDb failed: %w", err)
	}
	if err := srvApp.loadLoginAttemptsFromDb(ctx); err != nil {
		return fmt.Errorf("srvApp.loadLoginAttemptsFromDb failed: %w", err)
	}

	return nil
}

// startBackground starts the ws notification hub and the periodic maintenance goroutines.
// They stop when ctx is canceled.
func (srvApp *httpServerApplication) startBackground(ctx context.Context) (bg *background, err error) {

	// attach ws notification hub, listening on the same db channel found in system.notification_trigger()
	srvApp.NotificationHub, err = lysws.NewNotificationHub(ctx, srvApp.Db, "system.notification", maxUserWsConnections,
		srvApp.Config.UI.Url, srvApp.Logger)
	if err != nil {
		return nil, fmt.Errorf("lysws.NewNotificationHub failed: %w", err)
	}

	bg = &background{hubDone: make(chan struct{})}

	// start hub listener on its own lifecycle so hub failures do not stop the HTTP server
	go func() {
		defer close(bg.hubDone)
		if err := srvApp.NotificationHub.ListenAndBroadcast(ctx, sysnotification.SelectDetailsById); err != nil && !errors.Is(err, context.Canceled) {
			srvApp.Logger.Error("srvApp.NotificationHub.ListenAndBroadcast failed", "error", err)
		}
	}()

	// archive expired sessions immediately, and start goroutine to periodically archive them
	srvApp.archiveExpiredSessions(ctx)
	bg.wg.Go(func() { srvApp.runExpiredSessionArchiver(ctx, 5*time.Minute) })

	// start goroutine to periodically clean up expired rate limit buckets
	bg.wg.Go(func() { srvApp.runRateLimitCleanup(ctx, 30*time.Minute) })

	return bg, nil
}

// serve loads state, starts the background goroutines and serves on ln until ctx is canceled or the server fails.
// It then shuts everything down and returns the collected errors. stop must cancel ctx.
func (srvApp *httpServerApplication) serve(ctx context.Context, stop context.CancelFunc, srv *http.Server, ln net.Listener) error {

	defer ln.Close()

	if err := srvApp.loadState(ctx); err != nil {
		return err
	}

	bg, err := srvApp.startBackground(ctx)
	if err != nil {
		return err
	}

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting refsrv on port: %s", srvApp.Config.Api.Port)
	if srvApp.Config.General.Debug {
		startupMsg += ", debug: true"
	}
	srvApp.Logger.Info(startupMsg)

	// serve in a goroutine so that it doesn't block the main thread, which is waiting for shutdown signals
	srvErrCh := make(chan error, 1)
	go func() {
		srvErrCh <- srv.Serve(ln)
	}()

	// wait for shutdown signal or server error; hub failures are handled independently
	var serveErr error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = fmt.Errorf("srv.Serve failed: %w", err)
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	// stop background goroutines, restore default signal handling
	stop()

	return errors.Join(serveErr, srvApp.shutdown(srv, bg))
}

// shutdown stops everything started by serve. Every step runs regardless of earlier failures,
// and errors are collected and returned at the end. The ctx passed to startBackground must be canceled first.
func (srvApp *httpServerApplication) shutdown(srv *http.Server, bg *background) error {

	// wait for background goroutines to exit so none are mid-archive when sessions are persisted
	bg.wg.Wait()

	return errors.Join(
		srvApp.shutdownHttp(srv),
		srvApp.persistState(),
		srvApp.shutdownHub(bg.hubDone),
	)
}

// shutdownHttp gracefully shuts down the server, waiting for active requests to finish or timeout before forcefully closing
func (srvApp *httpServerApplication) shutdownHttp(srv *http.Server) error {

	ctx, cancel := context.WithTimeout(context.Background(), srvApp.ShutdownTimeouts.Http)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	// graceful shutdown failed (typically deadline exceeded): force-close remaining connections
	if closeErr := srv.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("srv.Close failed: %w", closeErr))
	}

	return fmt.Errorf("shutdown: srv.Shutdown failed: %w", err)
}

// persistState persists sessions and login attempts to db, for loading on restart
func (srvApp *httpServerApplication) persistState() error {

	ctx, cancel := context.WithTimeout(context.Background(), srvApp.ShutdownTimeouts.Persist)
	defer cancel()

	var errs error
	if err := srvApp.persistSessions(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("shutdown: srvApp.persistSessions failed: %w", err))
	}
	if err := srvApp.persistLoginAttempts(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("shutdown: srvApp.persistLoginAttempts failed: %w", err))
	}

	return errs
}

// shutdownHub waits for the hub listener to exit or timeout, then closes the hub
func (srvApp *httpServerApplication) shutdownHub(hubDone <-chan struct{}) error {

	var errs error

	select {
	case <-hubDone:
		// hub already stopped or exited on its own
	case <-time.After(srvApp.ShutdownTimeouts.Hub):
		errs = errors.Join(errs, errors.New("shutdown: timeout waiting for hub listener to exit"))
	}

	if err := srvApp.NotificationHub.Close(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("shutdown: srvApp.NotificationHub.Close failed: %w", err))
	}

	return errs
}
