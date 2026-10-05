package rootcli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/sql/ddl"
	"github.com/loveyourstack/lys-ref/sql/migrations"
	"github.com/loveyourstack/lys/lysos"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

func EnhanceMigrationFileCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "mig [migration file]",
		Short: "Replaces DDL asset names with their content in the migration file, and writes the result to the clipboard.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			res, err := lyspgdb.EnhanceMigrationFile(args[0], ddl.SQLAssets, migrations.SQLAssets)
			if err != nil {
				return err
			}

			// write to clipboard
			err = lysos.WriteToClipboard(res)
			if err != nil {
				return fmt.Errorf("lysos.WriteToClipboard failed: %w", err)
			}

			fmt.Println("success: written to clipboard")

			return nil
		},
	}
}
