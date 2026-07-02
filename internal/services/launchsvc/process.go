package launchsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
)

func (svc Service) ProcessFbLaunchItem(ctx context.Context, tx pgx.Tx, item dmlaunchfb.Model) (err error) {

	// status must be Queued
	if item.Status != launchstatus.Queued {
		return fmt.Errorf("status must be Queued, but is %s", item.Status)
	}

	// set status to In progress
	err = svc.FbLaunchStore.SetStatusTx(ctx, tx, launchstatus.InProgress, item.Id)
	if err != nil {
		return fmt.Errorf("svc.FbLaunchStore.SetStatusTx (InProgress) failed: %w", err)
	}

	// step 1: create creative unless already created
	if item.FbCreativeId == "" {

		// step must be 0
		if item.Step != 0 {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailedTx(ctx, tx, fmt.Sprintf("inconsistent state: step must be 0, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailedTx (Creative) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create creative
		creativeId, err := fakeFbApiCreateCreative(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailedTx (Creative) failed: %w", err2)
			}
			return nil
		}

		// set creative id and step to 1
		item.FbCreativeId = creativeId
		item.Step = 1
		err = svc.FbLaunchStore.SetCreativeIdTx(ctx, tx, creativeId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetCreativeIdTx failed: %w", err)
		}
	}

	// step 2: create campaign unless already created
	if item.FbCampaignId == "" {

		// step must be 1
		if item.Step != 1 {
			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailedTx(ctx, tx, fmt.Sprintf("inconsistent state: step must be 1, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailedTx (Campaign) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create campaign
		campaignId, err := fakeFbApiCreateCampaign(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailedTx (Campaign) failed: %w", err2)
			}
			return nil
		}

		// set campaign id and step to 2
		item.FbCampaignId = campaignId
		item.Step = 2
		err = svc.FbLaunchStore.SetCampaignIdTx(ctx, tx, campaignId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetCampaignIdTx failed: %w", err)
		}
	}

	// set status to Completed
	err = svc.FbLaunchStore.SetStatusTx(ctx, tx, launchstatus.Completed, item.Id)
	if err != nil {
		return fmt.Errorf("svc.FbLaunchStore.SetStatusTx (Completed) failed: %w", err)
	}

	return nil
}

func (svc Service) ProcessGAdsLaunchItem(ctx context.Context, tx pgx.Tx, item dmlaunchgads.Model) (err error) {

	// status must be Queued
	if item.Status != launchstatus.Queued {
		return fmt.Errorf("status must be Queued, but is %s", item.Status)
	}

	// set status to In progress
	err = svc.GAdsLaunchStore.SetStatusTx(ctx, tx, launchstatus.InProgress, item.Id)
	if err != nil {
		return fmt.Errorf("svc.GAdsLaunchStore.SetStatusTx (InProgress) failed: %w", err)
	}

	// step 1: create campaign unless already created
	if item.GAdsCampaignId == 0 {

		// step must be 0
		if item.Step != 0 {
			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, fmt.Sprintf("inconsistent state: step must be 0, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Campaign) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create campaign
		campaignId, err := fakeGAdsApiCreateCampaign(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Campaign) failed: %w", err2)
			}
			return nil
		}

		// set campaign id and step to 1
		item.GAdsCampaignId = campaignId
		item.Step = 1
		err = svc.GAdsLaunchStore.SetCampaignIdTx(ctx, tx, campaignId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetCampaignIdTx failed: %w", err)
		}
	}

	// step 2: create ad group unless already created
	if item.GAdsAdGroupId == 0 {

		// step must be 1
		if item.Step != 1 {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, fmt.Sprintf("inconsistent state: step must be 1, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Ad Group) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create ad group
		adGroupId, err := fakeGAdsApiCreateAdGroup(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Ad Group) failed: %w", err2)
			}
			return nil
		}

		// set ad group id and step to 2
		item.GAdsAdGroupId = adGroupId
		item.Step = 2
		err = svc.GAdsLaunchStore.SetAdGroupIdTx(ctx, tx, adGroupId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetAdGroupIdTx failed: %w", err)
		}
	}

	// step 3: create ad unless already created
	if item.GAdsAdId == 0 {

		// step must be 2
		if item.Step != 2 {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, fmt.Sprintf("inconsistent state: step must be 2, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Ad) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create ad
		adId, err := fakeGAdsApiCreateAd(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailedTx(ctx, tx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailedTx (Ad) failed: %w", err2)
			}
			return nil
		}

		// set ad id and step to 3
		item.GAdsAdId = adId
		item.Step = 3
		err = svc.GAdsLaunchStore.SetAdIdTx(ctx, tx, adId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetAdIdTx failed: %w", err)
		}
	}

	// set status to Completed
	err = svc.GAdsLaunchStore.SetStatusTx(ctx, tx, launchstatus.Completed, item.Id)
	if err != nil {
		return fmt.Errorf("svc.GAdsLaunchStore.SetStatusTx (Completed) failed: %w", err)
	}

	return nil
}
