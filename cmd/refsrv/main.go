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

	"github.com/loveyourstack/connectors/aws/awsapi"
	"github.com/loveyourstack/connectors/aws/awssvc"
	"github.com/loveyourstack/connectors/aws/stores/awsusersgrule"
	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys-ref/internal/services/syssvc"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys-ref/pkg/lysws"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
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

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		PostOptions: lys.FillPostOptions(lys.PostOptions{}), // use defaults

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3
	}

	// start goroutine to periodically clean up expired rate limit buckets
	go srvApp.runRateLimitCleanup(ctx, 30*time.Minute)

	// attach Get options
	srvApp.GetOptions, err = lys.FillGetOptions(lys.GetOptions{
		CsvDelimiter: '|',
	}) // use defaults for others
	if err != nil {
		log.Fatalf("initialization: failed to fill GetOptions: %s", err.Error())
	}

	// attach auth
	if srvApp.Config.General.Env == appenv.Prod {
		srvApp.UseXForwardedFor = true
		srvApp.XForwardedForIdx = 0
	}
	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                                                     // block after 5 failed attempts
	srvApp.Sessions = lysauth.NewAppSessions(srvApp.Validate, 10*time.Hour, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx) // allow 10 hour sessions

	// connect to db and assign pool to srvApp
	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbServerUser, srvApp.Config.General.AppName+" Srv", myapp.TypesToRegister)
	if err != nil {
		log.Fatalf("initialization: failed to create regular db connection pool: %s", err.Error())
	}
	defer srvApp.Db.Close()

	// attach ws notification hub to srvApp
	srvApp.NotificationHub, err = lysws.NewNotificationHub(ctx, srvApp.Db, "system.notification", srvApp.ErrorLog)
	if err != nil {
		log.Fatalf("initialization: failed to create notification hub: %s", err.Error())
	}
	defer srvApp.NotificationHub.Close()

	// start ws listener
	go func() {
		err := srvApp.NotificationHub.ListenAndBroadcast(ctx, sysnotification.SelectDetailsById)
		if err != nil {
			srvApp.ErrorLog.Error("srvApp.NotificationHub.ListenAndBroadcast failed", "error", err)
		}
	}()

	// attach stores
	srvApp.AwsUserSgRuleStore = awsusersgrule.Store{Db: srvApp.Db}
	srvApp.BlockedIPStore = sysblockedip.Store{Db: srvApp.Db}
	srvApp.GeoLocationStore = mmlocation.Store{Db: srvApp.Db}
	srvApp.GeoNetworkStore = mmnetwork.Store{Db: srvApp.Db}
	srvApp.LoginAttemptStore = sysloginattempt.Store{Db: srvApp.Db}
	srvApp.SessionStore = syssession.Store{Db: srvApp.Db}
	srvApp.SessionHistStore = syssessionhist.Store{Db: srvApp.Db}
	srvApp.SrvLogStore = syssrvreq.Store{Db: srvApp.Db}
	srvApp.UserStore = sysuser.Store{Db: srvApp.Db}

	// attach clients
	srvApp.AwsClient = awsapi.NewClient(conf.Aws, srvApp.Db, srvApp.InfoLog, srvApp.ErrorLog)

	// attach services
	srvApp.AwsSvc = awssvc.NewService(srvApp.Db, srvApp.AwsClient, srvApp.InfoLog, srvApp.ErrorLog)
	srvApp.ProcSvc = procsvc.NewService(conf.Process, srvApp.InfoLog, srvApp.ErrorLog)
	srvApp.SysSvc = syssvc.NewService(srvApp.InfoLog, srvApp.ErrorLog)

	// connect to db using db owner and assign to srvApp
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" Srv")
	if err != nil {
		log.Fatalf("initialization: failed to create owner db connection pool: %s", err.Error())
	}
	defer srvApp.OwnerDb.Close()

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

	// load persisted sessions and login attempts from db, if any
	err = srvApp.loadSessionsFromDb(ctx)
	if err != nil {
		log.Fatalf("initialization: srvApp.loadSessionsFromDb failed: %s", err.Error())
	}
	err = srvApp.loadLoginAttemptsFromDb(ctx)
	if err != nil {
		log.Fatalf("initialization: srvApp.loadLoginAttemptsFromDb failed: %s", err.Error())
	}

	// archive expired sessions immediately, and start goroutine to periodically archive them
	srvApp.archiveExpiredSessions(ctx)
	go srvApp.runExpiredSessionArchiver(ctx, 5*time.Minute)

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting refsrv on port: %s", srvApp.Config.API.Port)
	if conf.General.Debug {
		startupMsg += ", debug: true"
	}
	srvApp.InfoLog.Info(startupMsg)

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
		srvApp.InfoLog.Info("shutdown signal received")
	}

	// create context with timeout for server shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// gracefully shutdown server, waiting for active requests to finish or timeout before forcefully closing
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("shutdown: srv.Shutdown failed: %s", err.Error())
	}

	// persist sessions and login attempts to db, for loading on restart
	if err := srvApp.persistSessions(context.Background()); err != nil {
		srvApp.ErrorLog.Error("shutdown: srvApp.persistSessions failed", "error", err)
	}
	if err := srvApp.persistLoginAttempts(context.Background()); err != nil {
		srvApp.ErrorLog.Error("shutdown: srvApp.persistLoginAttempts failed", "error", err)
	}
}
