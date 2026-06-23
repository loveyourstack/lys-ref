package proccli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proc",
		Short: "Process schema commands",
	}

	cmd.AddCommand(MustRunCmd(cliApp))
	cmd.AddCommand(RunCmd(cliApp))
	cmd.AddCommand(RunOnlyCmd(cliApp))

	return cmd
}
