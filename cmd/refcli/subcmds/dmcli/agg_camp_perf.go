package dmcli

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperfagg"
	"github.com/spf13/cobra"
)

func AggCampPerfCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "aggCampPerf",
		Short: "Aggregate campaign performance data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			aggCampPerfStore := dmcampperfagg.Store{Db: cliApp.Db}
			err = aggCampPerfStore.Create(context.Background(), cliApp.InfoLog)
			if err != nil {
				return fmt.Errorf("aggCampPerfStore.Create failed: %w", err)
			}

			return nil
		},
	}
}
