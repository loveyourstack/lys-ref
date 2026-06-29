package rootcli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

func ResetDbDataCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "resetDbData",
		Short: "Deletes and recreates most database data.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()

			defer cliApp.Db.Close()

			dbOwnerUser := lyspgdb.User{
				Name:     cliApp.Config.DbOwnerUser.Name,
				Password: cliApp.Config.DbOwnerUser.Password,
			}

			// connect as db owner
			ownerDb, err := lyspgdb.GetPool(ctx, cliApp.Config.Db, dbOwnerUser, cliApp.Config.General.AppName+" (owner)")
			if err != nil {
				return fmt.Errorf("lyspgdb.GetPool (owner) failed: %w", err)
			}
			defer ownerDb.Close()

			filesToExecute := []string{
				"core/core_reset_tables.sql",
				"core/core_data.sql",

				"digmark/digmark_reset_tables.sql",
				"digmark/digmark_data.sql",

				"process/process_reset_tables.sql",
				"process/process_data.sql",

				"publisher/publisher_reset_tables.sql",
				"publisher/publisher_data.sql",
				"publisher/publisher_fake_updates.sql",

				"supplier/supplier_reset_tables.sql",
				"supplier/supplier_data.sql",

				"system/system_reset_tables.sql",
			}
			for _, file := range filesToExecute {
				err = lyspgdb.ExecuteFile(ctx, ownerDb, file, ddl.SQLAssets, nil, cliApp.Logger)
				if err != nil {
					return fmt.Errorf("lyspgdb.ExecuteFile failed for %s: %w", file, err)
				}
			}

			return nil
		},
	}
}
