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
	"github.com/loveyourstack/connectors/ecb/ecbapi"
	"github.com/loveyourstack/connectors/ecb/ecbsvc"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/connectors/maxmind/mmsvc"
	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/connectors/tedb/tedbapi"
	"github.com/loveyourstack/connectors/tedb/tedbsvc"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/connectors/awsbedrockapi"
	"github.com/loveyourstack/lys-ref/internal/connectors/gemapi"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys-ref/internal/services/syssvc"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysextdatasync"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lysws"
)

const maxUserWsConnections = 5

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

	// attach stores
	srvApp.AwsUserSgRuleStore = awsusersgrule.Store{Db: srvApp.Db}
	srvApp.BlockedIPStore = sysblockedip.Store{Db: srvApp.Db}
	srvApp.CountryStore = geocountry.Store{Db: srvApp.Db}
	srvApp.GeoLocationStore = mmlocation.Store{Db: srvApp.Db}
	srvApp.GeoNetworkStore = mmnetwork.Store{Db: srvApp.Db}
	srvApp.LoginAttemptStore = sysloginattempt.Store{Db: srvApp.Db}
	srvApp.SessionStore = syssession.Store{Db: srvApp.Db}
	srvApp.SessionHistStore = syssessionhist.Store{Db: srvApp.Db}
	srvApp.SrvLogStore = syssrvreq.Store{Db: srvApp.Db}
	srvApp.UserStore = sysuser.Store{Db: srvApp.Db}

	// attach clients
	srvApp.AwsClient = awsapi.NewClient(conf.Aws, srvApp.Db, srvApp.Logger)

	// bedrock: use separate client since it needs region override
	bedrockConf := awsapi.Conf{
		AccessKeyId:     conf.Aws.AccessKeyId,
		Region:          "us-west-2", // AWS Bedrock is only available in us-west-2 and us-east-1
		SecretAccessKey: conf.Aws.SecretAccessKey,
	}
	srvApp.AwsBedrockClient = awsbedrockapi.NewClient(bedrockConf, srvApp.Config.General.GeneratedPath, srvApp.Db, srvApp.Logger)

	srvApp.EcbClient = ecbapi.NewClient(srvApp.Db, srvApp.Logger)
	srvApp.GeminiClient = gemapi.NewClient(ctx, conf.Gemini, srvApp.Config.General.GeneratedPath, srvApp.Db, srvApp.Logger)
	srvApp.MaxMindClient = mmapi.NewClient(conf.MaxMind, srvApp.Db, srvApp.Logger)
	srvApp.TedbClient = tedbapi.NewClient(srvApp.Db, srvApp.Logger)

	// attach services
	syncStore := sysextdatasync.Store{Db: srvApp.Db}
	srvApp.AwsSvc = awssvc.NewService(srvApp.Db, srvApp.AwsClient, srvApp.Logger)
	srvApp.EcbSvc = ecbsvc.NewServiceWithSyncStore(srvApp.EcbClient, srvApp.Logger, syncStore)
	srvApp.MaxMindSvc = mmsvc.NewServiceWithSyncStore(srvApp.MaxMindClient, srvApp.Config.General.DownloadsPath, srvApp.Logger, syncStore)
	srvApp.ProcSvc = procsvc.NewService(conf.Process, srvApp.Logger)
	srvApp.SysSvc = syssvc.NewService(srvApp.Logger)
	srvApp.TedbSvc = tedbsvc.NewServiceWithSyncStore(srvApp.TedbClient, srvApp.Db, srvApp.Logger, syncStore)

	// connect to db using db owner and assign to srvApp
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" Srv")
	if err != nil {
		log.Fatalf("initialization: failed to create owner db connection pool: %s", err.Error())
	}
	defer srvApp.OwnerDb.Close()

	// create HTTP server using srvApp's routes and handlers
	rawHandler := srvApp.getRouter()
	timedHttpHandler := http.TimeoutHandler(rawHandler, 5*time.Second, "request timed out")

	srv := &http.Server{
		Addr: ":" + srvApp.Config.API.Port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// http.TimeoutHandler does not implement the http.Hijacker interface, which is needed for WebSocket upgrades
			// so allow websocket requests to skip the timeout handler
			if lysauth.IsWebSocket(r.Header) {
				rawHandler.ServeHTTP(w, r)
				return
			}

			// http requests
			timedHttpHandler.ServeHTTP(w, r)
		}),
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

	// --------------------------------

	// attach ws notification hub to srvApp, listening on the same db channel found in system.notification_trigger()
	srvApp.NotificationHub, err = lysws.NewNotificationHub(ctx, srvApp.Db, "system.notification", maxUserWsConnections,
		srvApp.Config.UI.Url, srvApp.Logger)
	if err != nil {
		log.Fatalf("initialization: failed to create notification hub: %s", err.Error())
	}

	// start hub listener on its own lifecycle so hub failures do not stop the HTTP server
	hubListenerDoneCh := make(chan struct{})
	go func() {
		defer close(hubListenerDoneCh)
		if err := srvApp.NotificationHub.ListenAndBroadcast(ctx, sysnotification.SelectDetailsById); err != nil && !errors.Is(err, context.Canceled) {
			srvApp.Logger.Error("srvApp.NotificationHub.ListenAndBroadcast failed", "error", err)
		}
	}()

	// --------------------------------

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting refsrv on port: %s", srvApp.Config.API.Port)
	if conf.General.Debug {
		startupMsg += ", debug: true"
	}
	srvApp.Logger.Info(startupMsg)

	// start server in a goroutine so that it doesn't block the main thread, which is waiting for shutdown signals
	srvErrCh := make(chan error, 1)
	go func() {
		srvErrCh <- srv.ListenAndServe()
	}()

	// wait for shutdown signal or server error; hub failures are handled independently
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvApp.Logger.Error("srv.ListenAndServe failed", "error", err)
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	// --------------------------------

	// create context with timeout for server shutdown
	srvShutdownCtx, srvCancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer srvCancelShutdown()

	// gracefully shutdown server, waiting for active requests to finish or timeout before forcefully closing
	if err := srv.Shutdown(srvShutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		srvApp.Logger.Error("shutdown: srv.Shutdown failed", "error", err)
	}

	// persist sessions and login attempts to db, for loading on restart
	if err := srvApp.persistSessions(context.Background()); err != nil {
		srvApp.Logger.Error("shutdown: srvApp.persistSessions failed", "error", err)
	}
	if err := srvApp.persistLoginAttempts(context.Background()); err != nil {
		srvApp.Logger.Error("shutdown: srvApp.persistLoginAttempts failed", "error", err)
	}

	// --------------------------------

	// gracefully shutdown notification hub, waiting for ListenAndBroadcast to exit or timeout before forcefully closing
	select {
	case <-hubListenerDoneCh:
		// exit: hub already stopped or exited on its own
	case <-time.After(5 * time.Second):
		srvApp.Logger.Error("shutdown: timeout waiting for hub listener to exit")
	}
	if err := srvApp.NotificationHub.Close(); err != nil {
		srvApp.Logger.Error("shutdown: srvApp.NotificationHub.Close failed", "error", err)
	}
}
