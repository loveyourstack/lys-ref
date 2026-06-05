package ecbsynccli

import (
	"fmt"
	"log/slog"

	"github.com/loveyourstack/lys-ref/cmd/refcli/cliapp"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurr"
	"github.com/spf13/cobra"
)

//	copied from ECB connector and modified so that currency metadata can be added

func CurrenciesCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "currencies",
		Short: "Sync currencies from ECB API into database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			defer cliApp.Db.Close()

			// select API items map with Code as key
			cApiItemsMap, err := cliApp.EcbClient.GetCurrenciesMap(cmd.Context())
			if err != nil {
				return fmt.Errorf("c.GetCurrenciesMap failed: %w", err)
			}

			// convert connector store type to local store type
			apiItemsMap := make(map[string]ecbcurr.Model)
			for key, apiItem := range cApiItemsMap {
				apiItemsMap[key] = ecbcurr.Model{
					Input: ecbcurr.Input{
						Code: apiItem.Code,
						Name: apiItem.Name,
					},
				}
			}

			itemStore := ecbcurr.Store{Db: cliApp.Db}
			itemType := "Currencies"

			// select DB items map with Code as key
			dbItemsMap, err := itemStore.SelectMapByNaturalKey(cmd.Context())
			if err != nil {
				return fmt.Errorf("itemStore.SelectMapByNaturalKey failed: %w", err)
			}

			// for each API item
			for key, apiItem := range apiItemsMap {

				// try to find the equivalent DB item
				dbItem, ok := dbItemsMap[key]
				if !ok {
					// insert to DB if not found
					_, err = itemStore.Insert(cmd.Context(), apiItem.Input)
					if err != nil {
						return fmt.Errorf("itemStore.Insert failed on key %v: %w", key, err)
					}
					cliApp.InfoLog.Info("inserted", slog.String("type", itemType), slog.Any("code", apiItem.Code))
					continue
				}

				// found: compare values and only update if needed
				if !itemStore.Equal(apiItem, dbItem) {

					err = itemStore.Update(cmd.Context(), apiItem.Input, dbItem.Id)
					if err != nil {
						return fmt.Errorf("itemStore.Update failed on key %v: %w", key, err)
					}
					cliApp.InfoLog.Info("updated", slog.String("type", itemType), slog.Any("code", apiItem.Code))
				}
			}

			// for each DB item
			for key, dbItem := range dbItemsMap {

				// try to find the equivalent API item
				_, ok := apiItemsMap[key]
				if !ok {
					// delete if not found
					err = itemStore.Delete(cmd.Context(), dbItem.Id)
					if err != nil {
						return fmt.Errorf("itemStore.Delete failed on key %v: %w", key, err)
					}
					cliApp.InfoLog.Info("deleted", slog.String("type", itemType), slog.Any("code", dbItem.Code))
				}
			}

			cliApp.InfoLog.Debug("done")

			return nil
		},
	}
}
