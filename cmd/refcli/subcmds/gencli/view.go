package gencli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys/lysgen"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

func ViewCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Generates PostgreSQL view definition from the supplied schema + table",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			// connect as owner (for permission to view constraints)
			dbOwnerUser := lyspgdb.User{
				Name:     cliApp.Config.DbOwnerUser.Name,
				Password: cliApp.Config.DbOwnerUser.Password,
			}

			ownerDb, err := lyspgdb.GetPool(cmd.Context(), cliApp.Config.Db, dbOwnerUser, cliApp.Config.General.AppName+" Cli")
			if err != nil {
				return fmt.Errorf("lyspgdb.GetPool (db owner) failed: %w", err)
			}
			defer ownerDb.Close()

			res, err := lysgen.View(cmd.Context(), ownerDb, args[0], args[1])
			if err != nil {
				return fmt.Errorf("lysgen.View failed: %w", err)
			}

			fmt.Println(res)

			return nil
		},
	}
}
