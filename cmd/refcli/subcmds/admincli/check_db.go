package admincli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lyspgmon"
	"github.com/spf13/cobra"
)

func CheckDbCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "checkDb",
		Short: "Checks database using lyspgmon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			// connect as owner (for permission to perform admin tasks)
			dbOwnerUser := lyspgdb.User{
				Name:     cliApp.Config.DbOwnerUser.Name,
				Password: cliApp.Config.DbOwnerUser.Password,
			}

			ownerDb, err := lyspgdb.GetPool(cmd.Context(), cliApp.Config.Db, dbOwnerUser, cliApp.Config.General.AppName+" Cli")
			if err != nil {
				return fmt.Errorf("lyspgdb.GetPool (db owner) failed: %w", err)
			}
			defer ownerDb.Close()

			err = lyspgmon.CheckDb(cmd.Context(), ownerDb, cliApp.InfoLog, cliApp.ErrorLog)
			if err != nil {
				return fmt.Errorf("lyspgmon.CheckDb failed: %w", err)
			}

			return nil
		},
	}
}
