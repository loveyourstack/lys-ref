package gencli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Code generation commands",
	}

	cmd.AddCommand(EqualCmd(cliApp))
	cmd.AddCommand(InputModelCmd(cliApp))
	cmd.AddCommand(TsTypesCmd(cliApp))
	cmd.AddCommand(ViewCmd(cliApp))

	return cmd
}
