package admincli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lyspgmon"
	"github.com/spf13/cobra"
)

func InstallLysPgMonCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "installLysPgMon",
		Short: "Adds lyspgmon schema, monitoring views and audit objects",
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

			// add schema, monitoring views and audit objects
			err = lyspgmon.Install(cmd.Context(), ownerDb, cliApp.Config.DbOwnerUser.Name, cliApp.Logger)
			if err != nil {
				return fmt.Errorf("lyspgmon.Install failed: %w", err)
			}

			// add lyspgmon permissions

			srvUserName := cliApp.Config.DbServerUser.Name
			cliUserName := cliApp.Config.DbCliUser.Name
			supplierUserName := "lysref_supplier" // external user not in config. Uses audit_update due to last_user_update_by col in supplier.product

			stmts := []string{}
			stmts = append(stmts, fmt.Sprintf("GRANT USAGE ON SCHEMA lyspgmon TO %s, %s, %s", srvUserName, cliUserName, supplierUserName))
			stmts = append(stmts, fmt.Sprintf("GRANT INSERT ON lyspgmon.audit_update TO %s, %s, %s", srvUserName, cliUserName, supplierUserName))
			stmts = append(stmts, fmt.Sprintf("GRANT SELECT ON ALL TABLES IN SCHEMA lyspgmon TO %s", srvUserName)) // includes views

			for i, stmt := range stmts {
				_, err = ownerDb.Exec(cmd.Context(), stmt)
				if err != nil {
					return fmt.Errorf("ownerDb.Exec (grant permissions) failed on stmt %d: %w", i+1, err)
				}
			}

			return nil
		},
	}
}
