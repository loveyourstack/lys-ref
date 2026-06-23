package pubcli

import (
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func NewCmd(cliApp *cliapp.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pub",
		Short: "Publisher commands",
	}

	cmd.AddCommand(AddFakeUpdatesCmd(cliApp))

	return cmd
}
