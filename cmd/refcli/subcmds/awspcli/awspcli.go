package awspcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS internal CLI commands",
	}

	cmd.AddCommand(ListSecGrpsCmd(cliApp))
	cmd.AddCommand(ListSecGrpRulesCmd(cliApp))

	return cmd
}
