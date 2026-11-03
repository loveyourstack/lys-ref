package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys/lystime"
	"google.golang.org/grpc"
)

// shutdownTimeouts are the maximum durations of the individual shutdown steps.
type shutdownTimeouts struct {
	Grpc    time.Duration // graceful gRPC shutdown, after which remaining connections are closed
	Persist time.Duration // persisting sessions and login attempts
}

var defaultShutdownTimeouts = shutdownTimeouts{
	Grpc:    10 * time.Second,
	Persist: 10 * time.Second,
}

type background struct {
	wg sync.WaitGroup
}

// loadState loads persisted sessions and login attempts from db, if any.
func (rpcApp *gRpcServerApplication) loadState(ctx context.Context) error {

	if err := rpcApp.loadSessionsFromDb(ctx); err != nil {
		return fmt.Errorf("rpcApp.loadSessionsFromDb failed: %w", err)
	}
	if err := rpcApp.loadLoginAttemptsFromDb(ctx); err != nil {
		return fmt.Errorf("rpcApp.loadLoginAttemptsFromDb failed: %w", err)
	}

	return nil
}

// startBackground starts the periodic maintenance goroutines. They stop when ctx is canceled.
func (rpcApp *gRpcServerApplication) startBackground(ctx context.Context) *background {

	bg := &background{}

	// archive expired sessions immediately, and start goroutine to periodically archive them
	rpcApp.archiveExpiredSessions(ctx)
	bg.wg.Go(func() {
		lystime.RunEvery(ctx, 5*time.Minute, func() { rpcApp.archiveExpiredSessions(ctx) })
	})

	// start goroutine to periodically clean up expired rate limit buckets
	bg.wg.Go(func() {
		lystime.RunEvery(ctx, 30*time.Minute, func() { rpcApp.AuthedRateLimits.CleanupExpired(); rpcApp.UnauthedRateLimits.CleanupExpired() })
	})

	return bg
}

// serve loads state, starts background goroutines and serves on ln until ctx is canceled or the server fails.
// It then shuts everything down and returns the collected errors. stop must cancel ctx.
func (rpcApp *gRpcServerApplication) serve(ctx context.Context, stop context.CancelFunc, grpcServer *grpc.Server, ln net.Listener) error {

	defer ln.Close()

	if err := rpcApp.loadState(ctx); err != nil {
		return err
	}

	bg := rpcApp.startBackground(ctx)

	srvErrCh := make(chan error, 1)
	go func() {
		cmd.StartupMsg("rpcsrv", rpcApp.Config.RpcApi.Port, rpcApp.Config.General.Debug, rpcApp.Logger)
		srvErrCh <- grpcServer.Serve(ln)
	}()

	var serveErr error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr = fmt.Errorf("grpcServer.Serve failed: %w", err)
		}
	case <-ctx.Done():
		rpcApp.Logger.Info("shutdown signal received")
	}

	stop()

	return errors.Join(serveErr, rpcApp.shutdown(grpcServer, bg))
}

// shutdown stops everything started by serve. The context passed to startBackground must be canceled first.
func (rpcApp *gRpcServerApplication) shutdown(grpcServer *grpc.Server, bg *background) error {

	bg.wg.Wait()

	return errors.Join(
		rpcApp.shutdownGrpc(grpcServer),
		rpcApp.persistState(),
	)
}

// shutdownGrpc gracefully shuts down the server, forcing it to stop when the timeout expires.
func (rpcApp *gRpcServerApplication) shutdownGrpc(grpcServer *grpc.Server) error {

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	err := lystime.WaitOrTimeout(stopped, rpcApp.ShutdownTimeouts.Grpc, "shutdown: graceful stop timed out, forcing stop")
	if err != nil {
		grpcServer.Stop()
	}

	return err
}

// persistState persists sessions and login attempts to db, for loading on restart.
func (rpcApp *gRpcServerApplication) persistState() error {

	ctx, cancel := context.WithTimeout(context.Background(), rpcApp.ShutdownTimeouts.Persist)
	defer cancel()

	var errs error
	if err := rpcApp.persistSessions(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("shutdown: rpcApp.persistSessions failed: %w", err))
	}
	if err := rpcApp.persistLoginAttempts(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("shutdown: rpcApp.persistLoginAttempts failed: %w", err))
	}

	return errs
}
