package fakecli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/fakecli/fakeinvcli"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fake",
		Short: "Fake commands for testing",
	}

	cmd.AddCommand(fakeinvcli.NewCmd(cliApp))

	cmd.AddCommand(AggCostsCmd(cliApp))
	cmd.AddCommand(AggRevenueCmd(cliApp))

	cmd.AddCommand(CreateReportsCmd(cliApp))

	cmd.AddCommand(FetchCostSource1Cmd(cliApp))
	cmd.AddCommand(FetchCostSource2Cmd(cliApp))
	cmd.AddCommand(FetchCostSource3Cmd(cliApp))

	cmd.AddCommand(FetchRevSource1Cmd(cliApp))
	cmd.AddCommand(FetchRevSource2Cmd(cliApp))

	cmd.AddCommand(FetchXRatesSourceCmd(cliApp))

	return cmd
}
