package tedbsynccli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "TODO",
	}

	cmd.AddCommand(VatRatesCmd(cliApp))

	return cmd
}
