package rootcli

import (
	"context"
	"fmt"

	"github.com/loveyourstack/connectors/aws/awsapi"
	"github.com/loveyourstack/connectors/aws/awssvc"
	"github.com/loveyourstack/connectors/ecb/ecbapi"
	"github.com/loveyourstack/connectors/ecb/ecbsvc"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/connectors/maxmind/mmsvc"
	"github.com/loveyourstack/connectors/tedb/tedbapi"
	"github.com/loveyourstack/connectors/tedb/tedbsvc"
	appCmd "github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysextdatasync"
	"github.com/loveyourstack/lys/lyspgdb"
)

// newCliApplication builds the application and connects everything it depends on.
// On error, any opened database pool is closed.
func newCliApplication(ctx context.Context, conf *myapp.Config) (cliApp *cliapp.App, err error) {

	cliApp = &cliapp.App{
		Application: appCmd.NewApplication(conf),
	}

	defer func() {
		if err != nil && cliApp.Db != nil {
			cliApp.Db.Close()
		}
	}()

	cliApp.Db, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbCliUser, conf.General.AppName+" Cli")
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPool failed: %w", err)
	}

	if err = attachClients(cliApp); err != nil {
		return nil, err
	}
	if err = attachServices(cliApp); err != nil {
		return nil, err
	}

	return cliApp, nil
}

func attachClients(cliApp *cliapp.App) (err error) {

	cliApp.AwsClient, err = awsapi.NewClient(cliApp.Config.Aws, cliApp.Db, cliApp.Logger)
	if err != nil {
		return fmt.Errorf("awsapi.NewClient failed: %w", err)
	}

	cliApp.EcbClient = ecbapi.NewClient(cliApp.Db, cliApp.Logger)

	cliApp.MaxMindClient, err = mmapi.NewClient(cliApp.Config.MaxMind, cliApp.Db, cliApp.Logger)
	if err != nil {
		return fmt.Errorf("mmapi.NewClient failed: %w", err)
	}

	cliApp.TedbClient = tedbapi.NewClient(cliApp.Db, cliApp.Logger)

	return nil
}

func attachServices(cliApp *cliapp.App) (err error) {

	syncStore := sysextdatasync.Store{Db: cliApp.Db}

	cliApp.AwsSvc, err = awssvc.NewService(cliApp.Db, cliApp.AwsClient, cliApp.Logger)
	if err != nil {
		return fmt.Errorf("awssvc.NewService failed: %w", err)
	}

	cliApp.EcbSvc = ecbsvc.NewServiceWithSyncStore(cliApp.EcbClient, cliApp.Logger, syncStore)

	cliApp.MaxMindSvc, err = mmsvc.NewServiceWithSyncStore(cliApp.MaxMindClient, cliApp.Config.General.DownloadsPath, cliApp.Logger, syncStore)
	if err != nil {
		return fmt.Errorf("mmsvc.NewServiceWithSyncStore failed: %w", err)
	}

	cliApp.ProcSvc, err = procsvc.NewService(cliApp.Config.Process, cliApp.Logger)
	if err != nil {
		return fmt.Errorf("procsvc.NewService failed: %w", err)
	}

	cliApp.TedbSvc = tedbsvc.NewServiceWithSyncStore(cliApp.TedbClient, cliApp.Db, cliApp.Logger, syncStore)

	return nil
}
