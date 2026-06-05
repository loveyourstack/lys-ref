package admincli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin commands",
	}

	cmd.AddCommand(CheckDbCmd(cliApp))
	cmd.AddCommand(InstallLysPgMonCmd(cliApp))

	return cmd
}
