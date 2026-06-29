package fakecli

import (
	"context"
	"errors"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func FetchCostSource1Cmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "costsource1 [from] [until]",
		Short: "Fetch cost source 1",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			err = cliApp.ProcSvc.RunFakeCmd(cmd.Context(), "costsource1", 3, cliApp.Logger, args[0], args[1])
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				return fmt.Errorf("cliApp.ProcSvc.RunFakeCmd failed: %w", err)
			}

			return nil
		},
	}
}
