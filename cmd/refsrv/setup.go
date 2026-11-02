package main

import (
	"context"
	"fmt"
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
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// newHttpServerApplication builds the application and connects everything it depends on.
// On error, anything already opened is closed. On success, the caller must call Close.
func newHttpServerApplication(ctx context.Context, conf *myapp.Config) (srvApp *httpServerApplication, err error) {

	srvApp = &httpServerApplication{
		Application: cmd.NewApplication(conf),
		PostOptions: lysapi.FillPostOptions(lysapi.PostOptions{}), // use defaults

		AuthedRateLimits:   lysauth.NewAppRateLimits(5, 20, 1*time.Hour),  // 5 rps with burst of 20. Buckets expire after 1 hour of inactivity
		UnauthedRateLimits: lysauth.NewAppRateLimits(0.5, 3, 1*time.Hour), // ~1 every 2s with burst of 3

		RpcApiKey:       conf.RpcClient.ApiKey,
		UseRealIpHeader: conf.General.Env == appenv.Prod,

		ShutdownTimeouts: defaultShutdownTimeouts,
	}

	defer func() {
		if err != nil {
			srvApp.Close()
			srvApp = nil
		}
	}()

	srvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{
		CsvDelimiter: '|',
	}) // use defaults for others
	if err != nil {
		return nil, fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	srvApp.attachAuth()

	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbServerUser, conf.General.AppName+" Srv", myapp.TypesToRegister)
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPoolWithTypes failed: %w", err)
	}

	srvApp.attachStores()

	if err = srvApp.attachRpcClient(); err != nil {
		return nil, err
	}
	if err = srvApp.attachClients(ctx); err != nil {
		return nil, err
	}
	if err = srvApp.attachServices(); err != nil {
		return nil, err
	}

	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" Srv")
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPool (owner) failed: %w", err)
	}

	return srvApp, nil
}

// Close releases the db pools and RPC client. It is safe to call on a partially built app.
func (srvApp *httpServerApplication) Close() {
	if srvApp.rpcClient != nil {
		srvApp.rpcClient.Close()
	}
	if srvApp.OwnerDb != nil {
		srvApp.OwnerDb.Close()
	}
	if srvApp.Db != nil {
		srvApp.Db.Close()
	}
}

// attachAuth must be called after UseRealIpHeader is set
func (srvApp *httpServerApplication) attachAuth() {
	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5)                                           // block after 5 failed attempts
	srvApp.Sessions = lysauth.NewAppSessions(srvApp.Validate, 10*time.Hour, srvApp.UseRealIpHeader) // allow 10 hour sessions
}

func (srvApp *httpServerApplication) attachStores() {
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
}

// attachRpcClient makes a client for the internal RPC server (doesn't actually connect until first request)
func (srvApp *httpServerApplication) attachRpcClient() (err error) {

	srvApp.rpcClient, err = grpc.NewClient("127.0.0.1:"+srvApp.Config.RpcApi.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),             // loopback connection only
		grpc.WithPerRPCCredentials(rpcKeyCredentials{key: srvApp.RpcApiKey}), // attach API key to outgoing RPC requests
		grpc.WithChainUnaryInterceptor(
			rpcForwardIpUnaryInterceptor(),                    // forward client IP to RPC server
			rpcDefaultDeadlineUnaryInterceptor(5*time.Second), // set default call timeout to 5 seconds
		),
	)
	if err != nil {
		return fmt.Errorf("grpc.NewClient failed: %w", err)
	}

	return nil
}

func (srvApp *httpServerApplication) attachClients(ctx context.Context) (err error) {

	conf := srvApp.Config

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

	return nil
}

func (srvApp *httpServerApplication) attachServices() (err error) {

	conf := srvApp.Config
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

	return nil
}
