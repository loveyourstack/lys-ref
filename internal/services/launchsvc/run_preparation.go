package launchsvc

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunch"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
)

// RunFbPreparation runs the preparation process for all unchecked Facebook launch items.
func (svc Service) RunFbPreparation(ctx context.Context) (err error) {

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select unchecked items to prepare, locking the selected rows
	items, err := svc.FbLaunchStore.SelectUncheckedTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("svc.FbLaunchStore.SelectUncheckedTx failed: %w", err)
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// select helpers
	helpers, err := svc.selectPrepareFbHelpers(ctx, tx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareFbHelpers failed: %w", err)
	}

	// for each item
	for _, item := range items {

		// parse campaign name, set unprepared if invalid
		parts, err := dmlaunchfb.ParseCampaignName(item.Name)
		if err != nil {
			err2 := svc.FbLaunchStore.SetUnpreparedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetUnpreparedTx (parse camp name) failed: %w", err2)
			}
			continue
		}

		// prepare item, set unprepared if invalid
		result, err := prepareFb(helpers, prepareFbParams{
			Account: parts.Account,
			prepareParams: prepareParams{
				CountryIso2: parts.CountryIso2,
				Vertical:    parts.Vertical,
			},
		})
		if err != nil {
			err2 := svc.FbLaunchStore.SetUnpreparedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetUnpreparedTx (prepare item) failed: %w", err2)
			}
			continue
		}

		// success: set prepared
		comp := dmlaunchfb.Computed{
			Computed: dmlaunch.Computed{
				CountryFk:  result.CountryFk,
				VerticalFk: result.VerticalFk,
			},
			FbAccountId: result.FbAccountId,
		}
		err = svc.FbLaunchStore.SetPreparedTx(ctx, tx, comp, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetPreparedTx failed: %w", err)
		}
	}

	// commit tx
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("svc.Db.Commit failed: %w", err)
	}

	return nil
}

// RunGAdsPreparation runs the preparation process for all unchecked GAds launch items.
func (svc Service) RunGAdsPreparation(ctx context.Context) (err error) {

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select unchecked items to prepare, locking the selected rows
	items, err := svc.GAdsLaunchStore.SelectUncheckedTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("svc.GAdsLaunchStore.SelectUncheckedTx failed: %w", err)
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// select helpers
	helpers, err := svc.selectPrepareGAdsHelpers(ctx, tx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareGAdsHelpers failed: %w", err)
	}

	// for each item
	for _, item := range items {

		// parse campaign name, set unprepared if invalid
		parts, err := dmlaunchgads.ParseCampaignName(item.Name)
		if err != nil {
			err2 := svc.GAdsLaunchStore.SetUnpreparedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetUnpreparedTx (parse camp name) failed: %w", err2)
			}
			continue
		}

		// prepare item, set unprepared if invalid
		result, err := prepareGAds(helpers, prepareGAdsParams{
			Account: parts.Account,
			prepareParams: prepareParams{
				CountryIso2: parts.CountryIso2,
				Vertical:    parts.Vertical,
			},
		})
		if err != nil {
			err2 := svc.GAdsLaunchStore.SetUnpreparedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetUnpreparedTx (prepare item) failed: %w", err2)
			}
			continue
		}

		// success: set prepared
		comp := dmlaunchgads.Computed{
			Computed: dmlaunch.Computed{
				CountryFk:  result.CountryFk,
				VerticalFk: result.VerticalFk,
			},
			GAdsAccountId: result.GAdsAccountId,
		}
		err = svc.GAdsLaunchStore.SetPreparedTx(ctx, tx, comp, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetPreparedTx failed: %w", err)
		}
	}

	// commit tx
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("svc.Db.Commit failed: %w", err)
	}

	return nil
}
