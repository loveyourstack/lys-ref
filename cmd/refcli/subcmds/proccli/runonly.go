package proccli

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func RunOnlyCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "runOnly stepId ...[key=value pairs]", // additional args, if sent, are key/value pairs in the form k=v
		Short: "TODO",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			stepId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("strconv.ParseInt failed: %w", err)
			}

			// create new run
			runId, err := cliApp.ProcSvc.CreateRunFromStep(cmd.Context(), cliApp.Db, stepId, args[1:], false)
			if err != nil {
				return fmt.Errorf("cliApp.ProcSvc.CreateRunFromStep failed: %w", err)
			}
			cliApp.InfoLog.Debug("created", slog.Int64("runId", runId))

			err = cliApp.ProcSvc.RunOnly(cmd.Context(), cliApp.Db, runId)
			if err != nil {
				return fmt.Errorf("cliApp.ProcSvc.RunOnly failed: %w", err)
			}

			return nil
		},
	}
}
