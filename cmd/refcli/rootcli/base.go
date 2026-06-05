package rootcli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/loveyourstack/connectors/ecb/ecbapi"
	"github.com/loveyourstack/connectors/ecb/ecbsvc"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/connectors/maxmind/mmsvc"
	appCmd "github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/admincli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/dmcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/ecbcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/fakecli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/gencli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/maxmindcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/proccli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/pubcli"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/procsvc"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:           "refcli",
	Version:       version,
	Short:         "refcli - CLI tool for LysRef",
	Long:          `refcli is a CLI tool for running LysRef admin tasks`,
	SilenceErrors: true, // subcommand errors are returned upwards via RunE and handled in Execute() below
	SilenceUsage:  true,
	// no Run function: a subcommand is always needed
}

var cliApp *cliapp.App

func addSubCommands() {
	rootCmd.AddCommand(CreateDbCmd(cliApp))
	rootCmd.AddCommand(GenHashCmd(cliApp))
	rootCmd.AddCommand(ResetDbDataCmd(cliApp))
	rootCmd.AddCommand(SleepCmd(cliApp))

	rootCmd.AddCommand(admincli.NewCmd(cliApp))
	rootCmd.AddCommand(dmcli.NewCmd(cliApp))
	rootCmd.AddCommand(ecbcli.NewCmd(cliApp))
	rootCmd.AddCommand(fakecli.NewCmd(cliApp))
	rootCmd.AddCommand(gencli.NewCmd(cliApp))
	rootCmd.AddCommand(maxmindcli.NewCmd(cliApp))
	rootCmd.AddCommand(proccli.NewCmd(cliApp))
	rootCmd.AddCommand(pubcli.NewCmd(cliApp))
}

func Execute() {

	configFileName := "ref_config.toml"

	// set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ensure that context cancelation propagates to subcommands
	rootCmd.SetContext(ctx)

	conf := myapp.Config{}
	err := conf.LoadFromFile(fmt.Sprintf("/usr/local/etc/%s", configFileName))
	if err != nil {
		log.Fatalf("initialization: conf.LoadFromFile (%s) failed: %s", configFileName, err.Error())
	}

	// create non-specific app
	app := appCmd.NewApplication(&conf)

	// create cli app
	cliApp = &cliapp.App{
		Application: app,
	}

	// connect to db and assign pool to cliApp
	cliApp.Db, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbCliUser, conf.General.AppName+" Cli")
	if err != nil {
		log.Fatalf("initialization: failed to create db connection pool: %s", err.Error())
	}
	defer cliApp.Db.Close()

	// attach API clients
	cliApp.EcbClient = ecbapi.NewClient(cliApp.Db, cliApp.InfoLog, cliApp.ErrorLog)
	cliApp.MaxMindClient = mmapi.NewClient(conf.MaxMind, cliApp.Db, cliApp.InfoLog, cliApp.ErrorLog)

	// attach services
	cliApp.EcbSvc = ecbsvc.NewService(cliApp.EcbClient, cliApp.InfoLog, cliApp.ErrorLog)
	cliApp.MaxMindSvc = mmsvc.NewService(cliApp.MaxMindClient, cliApp.Config.General.DownloadsPath, cliApp.InfoLog, cliApp.ErrorLog)
	cliApp.ProcSvc = procsvc.NewService(conf.Process, cliApp.InfoLog, cliApp.ErrorLog)

	// note that defer db Close is also needed in subcommands or else context cancelation doesn't propagate to db

	// subcommands
	addSubCommands()

	if err := rootCmd.Execute(); err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			log.Fatal(userErr)
		}
		log.Fatal(err.Error())
	}
}
