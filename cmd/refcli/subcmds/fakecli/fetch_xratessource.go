package fakecli

import (
	"context"
	"errors"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func FetchXRatesSourceCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "xratessource [freq]",
		Short: "Fetch exchange rates source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			err = cliApp.ProcSvc.RunFakeCmd(cmd.Context(), "xratessource", 2, cliApp.InfoLog, args[0])
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
