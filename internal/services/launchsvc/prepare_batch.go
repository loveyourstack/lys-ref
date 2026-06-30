package launchsvc

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunch"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
)

func (svc Service) PrepareFbBatch(ctx context.Context, batchSize int) (err error) {

	// select helpers
	helpers, err := svc.selectPrepareFbHelpers(ctx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareFbHelpers failed: %w", err)
	}

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select batch of unchecked items to prepare
	items, err := svc.FbLaunchStore.SelectPreparationBatch(ctx, tx, batchSize)
	if err != nil {
		return fmt.Errorf("svc.FbLaunchStore.SelectPreparationBatch failed: %w", err)
	}

	// for each item
	for _, item := range items {

		// parse campaign name, set unprepared if invalid
		parts, err := dmlaunchfb.ParseCampaignName(item.Name)
		if err != nil {
			err2 := svc.FbLaunchStore.SetUnprepared(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetUnprepared (parse camp name) failed: %w", err2)
			}
			continue
		}

		// prepare item, set unprepared if invalid
		result, err := PrepareFb(helpers, PrepareFbParams{
			Account: parts.Account,
			prepareParams: prepareParams{
				CountryIso2: parts.CountryIso2,
				Vertical:    parts.Vertical,
			},
		})
		if err != nil {
			err2 := svc.FbLaunchStore.SetUnprepared(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetUnprepared (prepare item) failed: %w", err2)
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
		err = svc.FbLaunchStore.SetPrepared(ctx, tx, comp, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetPrepared failed: %w", err)
		}
	}

	// commit tx
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("svc.Db.Commit failed: %w", err)
	}

	return nil
}

func (svc Service) PrepareGAdsBatch(ctx context.Context, batchSize int) (err error) {

	// select helpers
	helpers, err := svc.selectPrepareGAdsHelpers(ctx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareGAdsHelpers failed: %w", err)
	}

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select batch of unchecked items to prepare
	items, err := svc.GAdsLaunchStore.SelectPreparationBatch(ctx, tx, batchSize)
	if err != nil {
		return fmt.Errorf("svc.GAdsLaunchStore.SelectPreparationBatch failed: %w", err)
	}

	// for each item
	for _, item := range items {

		// parse campaign name, set unprepared if invalid
		parts, err := dmlaunchgads.ParseCampaignName(item.Name)
		if err != nil {
			err2 := svc.GAdsLaunchStore.SetUnprepared(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetUnprepared (parse camp name) failed: %w", err2)
			}
			continue
		}

		// prepare item, set unprepared if invalid
		result, err := PrepareGAds(helpers, PrepareGAdsParams{
			Account: parts.Account,
			prepareParams: prepareParams{
				CountryIso2: parts.CountryIso2,
				Vertical:    parts.Vertical,
			},
		})
		if err != nil {
			err2 := svc.GAdsLaunchStore.SetUnprepared(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetUnprepared (prepare item) failed: %w", err2)
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
		err = svc.GAdsLaunchStore.SetPrepared(ctx, tx, comp, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetPrepared failed: %w", err)
		}
	}

	// commit tx
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("svc.Db.Commit failed: %w", err)
	}

	return nil
}
