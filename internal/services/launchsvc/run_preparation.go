package launchsvc

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunch"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
)

func (svc Service) RunFbPreparation(ctx context.Context) (err error) {

	// select helpers
	helpers, err := svc.selectPrepareFbHelpers(ctx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareFbHelpers failed: %w", err)
	}

	for {

		// claim a batch of unprepared items
		items, err := svc.FbLaunchStore.ClaimForPreparation(ctx, svc.PrepBatchSize)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.ClaimForPreparation failed: %w", err)
		}

		// exit if none found
		if len(items) == 0 {
			break
		}

		// for each item
		for _, item := range items {

			// parse campaign name, set unprepared if invalid
			parts, err := dmlaunchfb.ParseCampaignName(item.Name)
			if err != nil {
				err2 := svc.FbLaunchStore.SetUnprepared(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.FbLaunchStore.SetUnprepared (parse camp name) failed: %w", err2)
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
				err2 := svc.FbLaunchStore.SetUnprepared(ctx, err.Error(), item.Id)
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
			err = svc.FbLaunchStore.SetPrepared(ctx, comp, item.Id)
			if err != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetPrepared failed: %w", err)
			}
		}
	}

	return nil
}

func (svc Service) RunGAdsPreparation(ctx context.Context) (err error) {

	// select helpers
	helpers, err := svc.selectPrepareGAdsHelpers(ctx)
	if err != nil {
		return fmt.Errorf("svc.selectPrepareGAdsHelpers failed: %w", err)
	}

	for {

		// claim a batch of unprepared items
		items, err := svc.GAdsLaunchStore.ClaimForPreparation(ctx, svc.PrepBatchSize)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.ClaimForPreparation failed: %w", err)
		}

		// exit if none found
		if len(items) == 0 {
			break
		}

		// for each item
		for _, item := range items {

			// parse campaign name, set unprepared if invalid
			parts, err := dmlaunchgads.ParseCampaignName(item.Name)
			if err != nil {
				err2 := svc.GAdsLaunchStore.SetUnprepared(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.GAdsLaunchStore.SetUnprepared (parse camp name) failed: %w", err2)
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
				err2 := svc.GAdsLaunchStore.SetUnprepared(ctx, err.Error(), item.Id)
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
			err = svc.GAdsLaunchStore.SetPrepared(ctx, comp, item.Id)
			if err != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetPrepared failed: %w", err)
			}
		}
	}

	return nil
}
