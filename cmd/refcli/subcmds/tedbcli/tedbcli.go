package tedbcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/tedbcli/tedbsynccli"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tedb",
		Short: "TEDB connector commands",
	}

	cmd.AddCommand(tedbsynccli.NewCmd(cliApp))

	return cmd
}
