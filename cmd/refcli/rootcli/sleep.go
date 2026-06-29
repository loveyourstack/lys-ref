package rootcli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/spf13/cobra"
)

func SleepCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sleep [sleepSecs] [cancelAfterSecs]",
		Short: "Sleeps for arg1 seconds, canceling after arg2 seconds. Set 0 for arg2 to disable cancellation.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// ensure that connection is closed even on context cancelation or other error
			defer cliApp.Db.Close()

			// only possible in dev env
			if cliApp.Config.General.Env != appenv.Dev {
				return fmt.Errorf("this command may only be used in dev environment")
			}

			sleepSecsStr := args[0]
			sleepSecs, err := strconv.Atoi(sleepSecsStr)
			if err != nil {
				return fmt.Errorf("strconv.Atoi failed: %w", err)
			}

			cancelAfterSecsStr := args[1]
			cancelAfterSecs, err := strconv.Atoi(cancelAfterSecsStr)
			if err != nil {
				return fmt.Errorf("strconv.Atoi failed: %w", err)
			}

			ctx := cmd.Context()

			if cancelAfterSecs > 0 && cancelAfterSecs < sleepSecs {
				cancelCtx, cancel := context.WithTimeout(ctx, time.Duration(cancelAfterSecs)*time.Second)
				ctx = cancelCtx
				defer cancel()

				go func() {
					<-ctx.Done()
					cliApp.Logger.Info(fmt.Sprintf("canceling sleep after %d seconds", cancelAfterSecs))
				}()
			}

			// execute db sleep query
			err = lyspg.Sleep(ctx, cliApp.Db, sleepSecs)
			if err != nil {
				return fmt.Errorf("lyspg.Sleep failed: %w", err)
			}

			return nil
		},
	}
}
