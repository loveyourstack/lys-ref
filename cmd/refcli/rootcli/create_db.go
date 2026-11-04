package rootcli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

// expects db users to have been created first (create_users.sql)

func CreateDbCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "createDb [test]",
		Short: "Creates database. Drops existing if present.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// only possible in dev env
			if cliApp.Config.General.Env != appenv.Dev {
				return fmt.Errorf("this command may only be used in dev environment")
			}

			// if there is a "test" argument, create the configured test database and add test data
			addTestData := false
			if len(args) == 1 && args[0] == "test" {
				if cliApp.Config.TestDatabase.Database == "" {
					return fmt.Errorf("test database is not configured")
				}
				addTestData = true
				cliApp.Config.Db.Database = cliApp.Config.TestDatabase.Database
			}

			// get file replacements from configs where needed
			fileReplacements, err := cliApp.Config.Developer.GetReplacements()
			if err != nil {
				return fmt.Errorf("cliApp.Config.Developer.GetReplacements failed: %w", err)
			}

			dbSuperUser := lyspgdb.User{
				Name:     cliApp.Config.DbSuperUser.Name,
				Password: cliApp.Config.DbSuperUser.Password,
			}

			params := lyspgdb.CreateLocalDbParams{
				SqlAssets:              ddl.SQLAssets,
				DbConf:                 cliApp.Config.Db,
				DbSuperUser:            dbSuperUser,
				DbOwnerConf:            cliApp.Config.DbOwnerUser,
				DropExisting:           true,
				AddSecurityPermissions: false,
				AddTestData:            addTestData,
				Replacements:           fileReplacements,
			}

			// (re-)create test db
			if err := lyspgdb.CreateLocalDb(cmd.Context(), params, cliApp.Logger); err != nil {
				return fmt.Errorf("lyspgdb.CreateLocalDb failed: %w", err)
			}

			return nil
		},
	}
}
