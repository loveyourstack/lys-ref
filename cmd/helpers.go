package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// NewHttpServer creates a new HTTP server with timeouts and optional bypass logic for specific requests.
func NewHttpServer(handler http.Handler, addr string, bypassTimeout func(*http.Request) bool) *http.Server {

	timedHandler := http.TimeoutHandler(handler, 5*time.Second, "request timed out")

	return &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// bypass timeout if needed
			if bypassTimeout != nil && bypassTimeout(r) {
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

// ShutdownHttp attempts to gracefully shut down the HTTP server within the specified timeout.
// If graceful shutdown fails, it force-closes remaining connections.
func ShutdownHttp(srv *http.Server, timeout time.Duration) error {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// attempt graceful shutdown with timeout
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

// StartupMsg logs a startup message with the server name, port, and debug status.
func StartupMsg(name, port string, debug bool, logger *slog.Logger) {
	startupMsg := fmt.Sprintf("starting %s on port: %s", name, port)
	if debug {
		startupMsg += ", debug: true"
	}
	logger.Info(startupMsg)
}
