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

			// try to parse campaign name
			parts, err := dmlaunchfb.ParseCampaignName(item.Name)
			if err != nil {
				err2 := svc.FbLaunchStore.SetInvalid(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.FbLaunchStore.SetInvalid (parse camp name) failed: %w", err2)
				}
				continue
			}

			// try to prepare item
			result, err := prepareFb(helpers, prepareFbParams{
				Account: parts.Account,
				prepareParams: prepareParams{
					CountryIso2: parts.CountryIso2,
					Vertical:    parts.Vertical,
				},
			})
			if err != nil {
				err2 := svc.FbLaunchStore.SetInvalid(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.FbLaunchStore.SetInvalid (prepare item) failed: %w", err2)
				}
				continue
			}

			// success: set ready
			comp := dmlaunchfb.Computed{
				Computed: dmlaunch.Computed{
					CountryFk:  result.CountryFk,
					VerticalFk: result.VerticalFk,
				},
				FbAccountId: result.FbAccountId,
			}
			err = svc.FbLaunchStore.SetReady(ctx, comp, item.Id)
			if err != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetReady failed: %w", err)
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

			// try to parse campaign name
			parts, err := dmlaunchgads.ParseCampaignName(item.Name)
			if err != nil {
				err2 := svc.GAdsLaunchStore.SetInvalid(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.GAdsLaunchStore.SetInvalid (parse camp name) failed: %w", err2)
				}
				continue
			}

			// try to prepare item
			result, err := prepareGAds(helpers, prepareGAdsParams{
				Account: parts.Account,
				prepareParams: prepareParams{
					CountryIso2: parts.CountryIso2,
					Vertical:    parts.Vertical,
				},
			})
			if err != nil {
				err2 := svc.GAdsLaunchStore.SetInvalid(ctx, err.Error(), item.Id)
				if err2 != nil {
					return fmt.Errorf("svc.GAdsLaunchStore.SetInvalid (prepare item) failed: %w", err2)
				}
				continue
			}

			// success: set ready
			comp := dmlaunchgads.Computed{
				Computed: dmlaunch.Computed{
					CountryFk:  result.CountryFk,
					VerticalFk: result.VerticalFk,
				},
				GAdsAccountId: result.GAdsAccountId,
			}
			err = svc.GAdsLaunchStore.SetReady(ctx, comp, item.Id)
			if err != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetReady failed: %w", err)
			}
		}
	}

	return nil
}
