package fakecli

import (
	"context"
	"errors"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func AggCostsCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "aggcosts [from] [until]",
		Short: "Aggregate costs",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			err = cliApp.ProcSvc.RunFakeCmd(cmd.Context(), "aggcosts", 8, cliApp.InfoLog, args[0], args[1])
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
