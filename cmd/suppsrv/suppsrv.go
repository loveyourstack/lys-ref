package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
)

func main() {
	configFilePath := flag.String("configFilePath", "suppsrv_config.toml", "Path to the config file")
	flag.Parse()

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

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		PostOptions: lysapi.FillPostOptions(lysapi.PostOptions{}), // use defaults

		RateLimits: lysauth.NewAppRateLimits(2, 5, 1*time.Hour), // 2 rps with burst of 5
	}

	// attach Get options
	srvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{}) // use defaults
	if err != nil {
		return fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	// attach auth
	if srvApp.Config.General.Env == appenv.Prod {
		srvApp.UseXForwardedFor = true
		srvApp.XForwardedForIdx = 0
	}
	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5) // block after 5 failed attempts

	// RLS connection (main connection for queries, with company_id setting applied for RLS)
	srvApp.Db, err = lyspgdb.GetPoolWithCtxSetting[int64](ctx, conf.Db, conf.DbServerUser, conf.General.AppName, "app.company_id", CompanyCtxKey, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPoolWithCtxSetting failed: %w", err)
	}
	defer srvApp.Db.Close()

	// owner db connection for queries that need to bypass RLS
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" (owner)")
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPool (owner) failed: %w", err)
	}
	defer srvApp.OwnerDb.Close()

	// attach stores
	srvApp.BlockedIPStore = sysblockedip.Store{Db: srvApp.OwnerDb}
	srvApp.SrvLogStore = syssrvreq.Store{Db: srvApp.OwnerDb}

	// create HTTP server using srvApp's routes and handlers
	srv := &http.Server{
		Addr:              ":" + srvApp.Config.Api.Port,
		Handler:           http.TimeoutHandler(srvApp.getRouter(), 5*time.Second, "request timed out"),
		IdleTimeout:       time.Second,
		MaxHeaderBytes:    1024 * 1024, // 1 MB
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
		WriteTimeout:      10 * time.Second,
	}

	// load persisted blocked IPs from db, if any
	err = srvApp.loadBlockedIPsFromDb(ctx)
	if err != nil {
		return fmt.Errorf("initialization: failed to load blocked IPs from db: %w", err)
	}

	// start goroutine to periodically clean up expired rate limit buckets
	go srvApp.runRateLimitCleanup(ctx, 30*time.Minute)

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting suppsrv on port: %s", conf.Api.Port)
	if conf.General.Debug {
		startupMsg += ", debug: true"
	}
	srvApp.Logger.Info(startupMsg)

	// start server in a goroutine so that it doesn't block the main thread, which is waiting for shutdown signals
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// wait for shutdown signal or server error
	var serveErr error
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = fmt.Errorf("srv.ListenAndServe failed: %w", err)
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	// stop background goroutines, restore default signal handling
	stop()

	// every shutdown step runs regardless of earlier failures; errors are collected and returned at the end
	runErr := serveErr

	// create context with timeout for server shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// gracefully shutdown server, waiting for active requests to finish or timeout before forcefully closing
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {

		// graceful shutdown failed (typically deadline exceeded): force-close remaining connections
		if closeErr := srv.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("srv.Close failed: %w", closeErr))
		}

		runErr = errors.Join(runErr, fmt.Errorf("shutdown: srv.Shutdown failed: %w", err))
	}

	return runErr
}
