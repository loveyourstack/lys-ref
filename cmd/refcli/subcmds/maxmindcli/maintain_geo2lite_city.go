package maxmindcli

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/connectors/maxmind/stores/mmgeodbmeta"
	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/spf13/cobra"
)

func MaintainGeo2LiteCityCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:  "maintainGeo2LiteCity",
		Long: "Checks if the GeoLite2 City data in the database is up to date and updates it if necessary.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()

			defer cliApp.Db.Close()

			// try to get last write time from db
			lastWriteAt := time.Time{}
			metaStore := mmgeodbmeta.Store{Db: cliApp.Db}
			meta, err := metaStore.SelectByGeoDb(ctx, string(mmapi.GeoLite2City))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// continue with zero value
				} else {
					return fmt.Errorf("metaStore.SelectByGeoDb failed: %w", err)
				}
			} else {
				lastWriteAt = meta.LastWriteAt.ToTime()
			}

			// get last modified time from API
			lastModified, err := cliApp.MaxMindClient.GetApiGeoLite2DbLastModified(ctx, mmapi.GeoLite2City)
			if err != nil {
				return fmt.Errorf("cliApp.MaxMindClient.GetApiGeoLite2DbLastModified failed: %w", err)
			}

			// exit if no update needed
			if lastWriteAt.After(lastModified) || lastWriteAt.Equal(lastModified) {
				cliApp.Logger.Debug("no GeoLite2 update needed", slog.Time("lastWriteAt", lastWriteAt), slog.Time("lastModified", lastModified))
				return nil
			}

			// update needed
			cliApp.Logger.Info("running GeoLite2 update", slog.Time("lastWriteAt", lastWriteAt), slog.Time("lastModified", lastModified))
			err = cliApp.MaxMindSvc.WriteGeo2LiteCity(ctx, cliApp.Db, true)
			if err != nil {
				return fmt.Errorf("cliApp.MaxMindSvc.WriteGeo2LiteCity failed: %w", err)
			}

			return nil
		},
	}
}
