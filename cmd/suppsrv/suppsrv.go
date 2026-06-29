package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
)

func main() {

	configFileName := "suppsrv_config.toml"

	// mandatory flag if not using default
	configFilePath := flag.String("configFilePath", configFileName, "Path to the config file")

	flag.Parse()

	// load config from file
	conf := myapp.Config{}
	err := conf.LoadFromFile(*configFilePath)
	if err != nil {
		log.Fatalf("initialization: conf.LoadFromFile (%s) failed: %s", *configFilePath, err.Error())
	}

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		PostOptions: lys.FillPostOptions(lys.PostOptions{}), // use defaults

		RateLimits: lysauth.NewAppRateLimits(2, 5, 1*time.Hour), // 2 rps with burst of 5
	}

	// start goroutine to periodically clean up expired rate limit buckets
	go srvApp.runRateLimitCleanup(ctx, 30*time.Minute)

	// attach Get options
	srvApp.GetOptions, err = lys.FillGetOptions(lys.GetOptions{}) // use defaults
	if err != nil {
		log.Fatalf("initialization: failed to fill GetOptions: %s", err.Error())
	}

	// attach auth
	if srvApp.Config.General.Env == appenv.Prod {
		srvApp.UseXForwardedFor = true
		srvApp.XForwardedForIdx = 0
	}
	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5) // block after 5 failed attempts

	// RLS connection (main connection for queries, with company_id setting applied for RLS)
	srvApp.Db, err = lyspgdb.GetPoolWithCtxSetting[int64](ctx, conf.Db, conf.DbServerUser, srvApp.Config.General.AppName, "app.company_id", CompanyCtxKey, srvApp.Logger)
	if err != nil {
		log.Fatalf("initialization: failed to create RLS db connection pool: %s", err.Error())
	}
	defer srvApp.Db.Close()

	// owner db connection for queries that need to bypass RLS
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" (owner)")
	if err != nil {
		log.Fatalf("initialization: failed to create owner db connection pool: %s", err.Error())
	}
	defer srvApp.OwnerDb.Close()

	// attach stores
	srvApp.BlockedIPStore = sysblockedip.Store{Db: srvApp.OwnerDb}
	srvApp.SrvLogStore = syssrvreq.Store{Db: srvApp.OwnerDb}

	// create HTTP server using srvApp's routes and handlers
	srv := &http.Server{
		Addr:              ":" + srvApp.Config.API.Port,
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
		log.Fatalf("initialization: failed to load blocked IPs from db: %s", err.Error())
	}

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting suppsrv on port: %s", srvApp.Config.API.Port)
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
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("initialization: srv.ListenAndServe failed: %s", err.Error())
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	// create context with timeout for server shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// gracefully shutdown server, waiting for active requests to finish or timeout before forcefully closing
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("shutdown: srv.Shutdown failed: %s", err.Error())
	}
}
