package pubcli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

func AddFakeUpdatesCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "addFakeUpdates",
		Short: "Add fake updates to the publisher schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			err = lyspgdb.ExecuteFile(cmd.Context(), cliApp.Db, "publisher/publisher_fake_updates.sql", ddl.SQLAssets, nil, cliApp.Logger)
			if err != nil {
				return fmt.Errorf("lyspgdb.ExecuteFile failed: %w", err)
			}

			return nil
		},
	}
}
