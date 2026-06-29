package fakeinvcli

import (
	"context"
	"errors"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func GenPdfsCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "genpdfs [month]",
		Short: "Generate invoice PDFs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			err = cliApp.ProcSvc.RunFakeCmd(cmd.Context(), "invoices genpdfs", 3, cliApp.Logger, args[0])
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
