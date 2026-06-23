package dmcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dm",
		Short: "Digital marketing commands",
	}

	cmd.AddCommand(AggCampPerfCmd(cliApp))

	return cmd
}
