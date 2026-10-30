package rootcli

import (
	"crypto/sha256"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys/lysstring"
	"github.com/spf13/cobra"
)

func GenKeyCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "genkey",
		Short: "Generates an API key and its hash.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// only possible in dev env
			if cliApp.Config.General.Env != appenv.Dev {
				return fmt.Errorf("this command may only be used in dev environment")
			}

			key := lysstring.Rand(32)

			h := sha256.New()
			h.Write([]byte(key))
			hashed := fmt.Sprintf("%x", h.Sum(nil))

			fmt.Println("key:", string(key))
			fmt.Println("hashed:", hashed)

			return nil
		},
	}
}
