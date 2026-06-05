package rootcli

import (
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func GenHashCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "genhash [input]",
		Short: "Generates a hash for the given input.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// ensure that connection is closed even on context cancelation or other error
			defer cliApp.Db.Close()

			// only possible in dev env
			if cliApp.Config.General.Env != appenv.Dev {
				return fmt.Errorf("this command may only be used in dev environment")
			}

			input := args[0]
			if len(input) == 0 {
				return fmt.Errorf("input missing")
			}

			hashed, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("bcrypt.GenerateFromPassword failed: %w", err)
			}

			fmt.Println(string(hashed))

			return nil
		},
	}
}
