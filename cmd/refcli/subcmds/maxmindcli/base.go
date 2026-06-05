package maxmindcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maxmind",
		Short: "MaxMind commands",
	}

	cmd.AddCommand(MaintainGeo2LiteCityCmd(cliApp))
	cmd.AddCommand(WriteGeo2LiteCityCmd(cliApp))

	return cmd
}
