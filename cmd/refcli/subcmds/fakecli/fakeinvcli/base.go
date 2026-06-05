package fakeinvcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Fake invoice commands for testing",
	}

	cmd.AddCommand(CreateDataCmd(cliApp))
	cmd.AddCommand(GenPdfsCmd(cliApp))
	cmd.AddCommand(SendCmd(cliApp))
	cmd.AddCommand(ValRecipsCmd(cliApp))

	return cmd
}
