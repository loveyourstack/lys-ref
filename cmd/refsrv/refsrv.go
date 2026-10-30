package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lysws"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const maxUserWsConnections = 5

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

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		PostOptions: lysapi.FillPostOptions(lysapi.PostOptions{}), // use defaults

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3

		RpcApiKey: conf.RpcClient.ApiKey,
	}

	// attach Get options
	srvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		CsvDelimiter: '|',
	}) // use defaults for others
	if err != nil {
		return fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	// attach auth
	if conf.General.Env == appenv.Prod {
		srvApp.UseXForwardedFor = true
		srvApp.XForwardedForIdx = 0
	}
	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                                                     // block after 5 failed attempts
	srvApp.Sessions = lysauth.NewAppSessions(srvApp.Validate, 10*time.Hour, srvApp.UseXForwardedFor, srvApp.XForwardedForIdx) // allow 10 hour sessions

	// connect to db and assign pool to srvApp
	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbServerUser, conf.General.AppName+" Srv", myapp.TypesToRegister)
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPoolWithTypes failed: %w", err)
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

	// make client for internal RPC server (doesn't actually connect until first request)
	srvApp.rpcClient, err = grpc.NewClient("localhost:"+conf.RpcApi.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(rpcKeyCredentials{key: srvApp.RpcApiKey}), // attach API key to outgoing RPC requests
		grpc.WithChainUnaryInterceptor(
			rpcForwardIpUnaryInterceptor(),                    // forward client IP to RPC server
			rpcDefaultDeadlineUnaryInterceptor(5*time.Second), // set default call timeout to 5 seconds
		),
	)
	if err != nil {
		return fmt.Errorf("grpc.NewClient failed: %w", err)
	}
	defer srvApp.rpcClient.Close()

	// attach clients
	srvApp.AwsClient, err = awsapi.NewClient(conf.Aws, srvApp.Db, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("awsapi.NewClient failed: %w", err)
	}

	// bedrock: use separate client since it needs region override
	bedrockConf := awsapi.Conf{
		AccessKeyId:     conf.Aws.AccessKeyId,
		Region:          "us-west-2", // AWS Bedrock is only available in us-west-2 and us-east-1
		SecretAccessKey: conf.Aws.SecretAccessKey,
	}
	srvApp.AwsBedrockClient, err = awsbedrockapi.NewClient(bedrockConf, conf.General.GeneratedPath, srvApp.Db, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("awsbedrockapi.NewClient failed: %w", err)
	}

	srvApp.EcbClient = ecbapi.NewClient(srvApp.Db, srvApp.Logger)

	srvApp.GeminiClient, err = gemapi.NewClient(ctx, conf.Gemini, conf.General.GeneratedPath, srvApp.Db, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("gemapi.NewClient failed: %w", err)
	}

	srvApp.MaxMindClient, err = mmapi.NewClient(conf.MaxMind, srvApp.Db, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("mmapi.NewClient failed: %w", err)
	}
	srvApp.TedbClient = tedbapi.NewClient(srvApp.Db, srvApp.Logger)

	// attach services
	syncStore := sysextdatasync.Store{Db: srvApp.Db}
	srvApp.AwsSvc, err = awssvc.NewService(srvApp.Db, srvApp.AwsClient, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("awssvc.NewService failed: %w", err)
	}
	srvApp.EcbSvc = ecbsvc.NewServiceWithSyncStore(srvApp.EcbClient, srvApp.Logger, syncStore)
	srvApp.MaxMindSvc, err = mmsvc.NewServiceWithSyncStore(srvApp.MaxMindClient, conf.General.DownloadsPath, srvApp.Logger, syncStore)
	if err != nil {
		return fmt.Errorf("mmsvc.NewServiceWithSyncStore failed: %w", err)
	}
	srvApp.ProcSvc, err = procsvc.NewService(conf.Process, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("procsvc.NewService failed: %w", err)
	}
	srvApp.SysSvc = syssvc.NewService(srvApp.AwsClient, conf.General.S3Bucket, srvApp.Db, srvApp.Logger)
	srvApp.TedbSvc = tedbsvc.NewServiceWithSyncStore(srvApp.TedbClient, srvApp.Db, srvApp.Logger, syncStore)

	// connect to db using db owner and assign to srvApp
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" Srv")
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPool (owner) failed: %w", err)
	}
	defer srvApp.OwnerDb.Close()

	// create HTTP server using srvApp's routes and handlers
	rawHandler := srvApp.getRouter()
	timedHttpHandler := http.TimeoutHandler(rawHandler, 5*time.Second, "request timed out")

	srv := &http.Server{
		Addr: ":" + conf.Api.Port,
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
		return fmt.Errorf("srvApp.loadSessionsFromDb failed: %w", err)
	}
	err = srvApp.loadLoginAttemptsFromDb(ctx)
	if err != nil {
		return fmt.Errorf("srvApp.loadLoginAttemptsFromDb failed: %w", err)
	}

	// --------------------------------

	// attach ws notification hub to srvApp, listening on the same db channel found in system.notification_trigger()
	srvApp.NotificationHub, err = lysws.NewNotificationHub(ctx, srvApp.Db, "system.notification", maxUserWsConnections,
		conf.UI.Url, srvApp.Logger)
	if err != nil {
		return fmt.Errorf("lysws.NewNotificationHub failed: %w", err)
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

	// create wait group for background goroutines that must exit before sessions are persisted
	var bgWg sync.WaitGroup

	// archive expired sessions immediately, and start goroutine to periodically archive them
	srvApp.archiveExpiredSessions(ctx)
	bgWg.Go(func() { srvApp.runExpiredSessionArchiver(ctx, 5*time.Minute) })

	// start goroutine to periodically clean up expired rate limit buckets
	bgWg.Go(func() { srvApp.runRateLimitCleanup(ctx, 30*time.Minute) })

	// display startup message with port and debug mode if enabled
	startupMsg := fmt.Sprintf("starting refsrv on port: %s", conf.Api.Port)
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
	var serveErr error
	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = fmt.Errorf("srv.ListenAndServe failed: %w", err)
		}
	case <-ctx.Done():
		srvApp.Logger.Info("shutdown signal received")
	}

	// stop background goroutines, restore default signal handling
	stop()

	// wait for background goroutines to exit so none are mid-archive when sessions are persisted
	bgWg.Wait()

	// every shutdown step runs regardless of earlier failures; errors are collected and returned at the end
	runErr := serveErr

	// --------------------------------

	// create context with timeout for server shutdown
	srvShutdownCtx, srvCancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer srvCancelShutdown()

	// gracefully shutdown server, waiting for active requests to finish or timeout before forcefully closing
	if err := srv.Shutdown(srvShutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {

		// graceful shutdown failed (typically deadline exceeded): force-close remaining connections
		if closeErr := srv.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("srv.Close failed: %w", closeErr))
		}

		runErr = errors.Join(runErr, fmt.Errorf("shutdown: srv.Shutdown failed: %w", err))
	}

	persistCtx, persistCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer persistCancel()

	// persist sessions and login attempts to db, for loading on restart
	if err := srvApp.persistSessions(persistCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: srvApp.persistSessions failed: %w", err))
	}
	if err := srvApp.persistLoginAttempts(persistCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: srvApp.persistLoginAttempts failed: %w", err))
	}

	// --------------------------------

	// gracefully shutdown notification hub, waiting for ListenAndBroadcast to exit or timeout before forcefully closing
	select {
	case <-hubListenerDoneCh:
		// exit: hub already stopped or exited on its own
	case <-time.After(5 * time.Second):
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: timeout waiting for hub listener to exit"))
	}
	if err := srvApp.NotificationHub.Close(); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown: srvApp.NotificationHub.Close failed: %w", err))
	}

	return runErr
}
