package rootcli

import (
	"fmt"
	"log"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

// expects db users to have been created first (create_users.sql)

func CreateDbCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "createDb",
		Short: "Creates database. Drops existing if present.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			// only possible in dev env
			if cliApp.Config.General.Env != appenv.Dev {
				return fmt.Errorf("this command may only be used in dev environment")
			}

			// get file replacements from configs where needed
			fileReplacements, err := cliApp.Config.Developer.GetReplacements()
			if err != nil {
				log.Fatalf("cliApp.Config.Developer.GetReplacements failed: %s", err.Error())
			}

			dbSuperUser := lyspgdb.User{
				Name:     cliApp.Config.DbSuperUser.Name,
				Password: cliApp.Config.DbSuperUser.Password,
			}

			// (re-)create test db
			if err := lyspgdb.CreateLocalDb(cmd.Context(), ddl.SQLAssets, cliApp.Config.Db, dbSuperUser, cliApp.Config.DbOwnerUser, true, false,
				fileReplacements, cliApp.InfoLog); err != nil {
				return fmt.Errorf("lyspgdb.CreateLocalDb failed: %w", err)
			}

			return nil
		},
	}
}
