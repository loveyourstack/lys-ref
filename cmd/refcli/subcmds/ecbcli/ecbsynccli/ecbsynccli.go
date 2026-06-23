package ecbsynccli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "TODO",
	}

	cmd.AddCommand(CurrenciesCmd(cliApp))
	cmd.AddCommand(ExchangeRatesCmd(cliApp))

	return cmd
}
