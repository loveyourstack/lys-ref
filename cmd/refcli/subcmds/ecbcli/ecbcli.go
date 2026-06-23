package ecbcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/cmd/refcli/subcmds/ecbcli/ecbsynccli"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecb",
		Short: "ECB connector commands",
	}

	cmd.AddCommand(ecbsynccli.NewCmd(cliApp))

	return cmd
}
