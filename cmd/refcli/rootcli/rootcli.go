package rootcli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/admincli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/awspcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/dmcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/ecbcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/fakecli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/gencli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/maxmindcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/proccli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/pubcli"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/tedbcli"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:           "refcli",
	Version:       version,
	Short:         "refcli - CLI tool for LysRef",
	Long:          `refcli is a CLI tool for running LysRef admin tasks`,
	SilenceErrors: true, // subcommand errors are returned upwards via RunE to Run() below, then to main()
	SilenceUsage:  true,
	// no Run function: a subcommand is always needed
}

func addSubCommands(cliApp *cliapp.App) {
	rootCmd.AddCommand(CreateDbCmd(cliApp))
	rootCmd.AddCommand(GenHashCmd(cliApp))
	rootCmd.AddCommand(GenKeyCmd(cliApp))
	rootCmd.AddCommand(EnhanceMigrationFileCmd(cliApp))
	rootCmd.AddCommand(ResetDbDataCmd(cliApp))
	rootCmd.AddCommand(SleepCmd(cliApp))

	rootCmd.AddCommand(admincli.NewCmd(cliApp))
	rootCmd.AddCommand(awspcli.NewCmd(cliApp))
	rootCmd.AddCommand(dmcli.NewCmd(cliApp))
	rootCmd.AddCommand(ecbcli.NewCmd(cliApp))
	rootCmd.AddCommand(fakecli.NewCmd(cliApp))
	rootCmd.AddCommand(gencli.NewCmd(cliApp))
	rootCmd.AddCommand(maxmindcli.NewCmd(cliApp))
	rootCmd.AddCommand(proccli.NewCmd(cliApp))
	rootCmd.AddCommand(pubcli.NewCmd(cliApp))
	rootCmd.AddCommand(tedbcli.NewCmd(cliApp))
}

func Run() error {

	configFileName := "ref_config.toml"

	// set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ensure that context cancelation propagates to subcommands
	rootCmd.SetContext(ctx)

	conf := myapp.Config{}
	err := conf.LoadFromFile(fmt.Sprintf("/usr/local/etc/%s", configFileName))
	if err != nil {
		return fmt.Errorf("conf.LoadFromFile (%s) failed: %w", configFileName, err)
	}

	// build and wire up the CLI application
	cliApp, err := newCliApplication(ctx, &conf)
	if err != nil {
		return err
	}
	defer cliApp.Db.Close()

	// add subcommands
	addSubCommands(cliApp)

	// execute the root command, which runs the appropriate subcommand based on user input
	return rootCmd.Execute()
}
