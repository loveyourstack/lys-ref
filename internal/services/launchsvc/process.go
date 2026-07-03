package launchsvc

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
)

func (svc Service) ProcessFbLaunchItem(ctx context.Context, item dmlaunchfb.Model) (err error) {

	// status must be Processing
	if item.Status != launchstatus.Processing {
		return fmt.Errorf("status must be Processing, but is %s", item.Status)
	}

	// step 1: create creative unless already created
	if item.FbCreativeId == "" {

		// step must be 0
		if item.Step != 0 {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailed(ctx, fmt.Sprintf("inconsistent state: step must be 0, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailed (Creative) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create creative
		creativeId, err := fakeFbApiCreateCreative(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailed(ctx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailed (Creative) failed: %w", err2)
			}
			return nil
		}

		// set creative id and step to 1
		item.FbCreativeId = creativeId
		item.Step = 1
		err = svc.FbLaunchStore.SetCreativeId(ctx, creativeId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetCreativeId failed: %w", err)
		}
	}

	// step 2: create campaign unless already created
	if item.FbCampaignId == "" {

		// step must be 1
		if item.Step != 1 {
			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailed(ctx, fmt.Sprintf("inconsistent state: step must be 1, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailed (Campaign) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create campaign
		campaignId, err := fakeFbApiCreateCampaign(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.FbLaunchStore.SetFailed(ctx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.FbLaunchStore.SetFailed (Campaign) failed: %w", err2)
			}
			return nil
		}

		// set campaign id and step to 2
		item.FbCampaignId = campaignId
		item.Step = 2
		err = svc.FbLaunchStore.SetCampaignId(ctx, campaignId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.FbLaunchStore.SetCampaignId failed: %w", err)
		}
	}

	// set status to Completed
	err = svc.FbLaunchStore.SetStatus(ctx, launchstatus.Completed, item.Id)
	if err != nil {
		return fmt.Errorf("svc.FbLaunchStore.SetStatus (Completed) failed: %w", err)
	}

	return nil
}

func (svc Service) ProcessGAdsLaunchItem(ctx context.Context, item dmlaunchgads.Model) (err error) {

	// status must be Processing
	if item.Status != launchstatus.Processing {
		return fmt.Errorf("status must be Processing, but is %s", item.Status)
	}

	// step 1: create campaign unless already created
	if item.GAdsCampaignId == 0 {

		// step must be 0
		if item.Step != 0 {
			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, fmt.Sprintf("inconsistent state: step must be 0, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Campaign) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create campaign
		campaignId, err := fakeGAdsApiCreateCampaign(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Campaign) failed: %w", err2)
			}
			return nil
		}

		// set campaign id and step to 1
		item.GAdsCampaignId = campaignId
		item.Step = 1
		err = svc.GAdsLaunchStore.SetCampaignId(ctx, campaignId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetCampaignId failed: %w", err)
		}
	}

	// step 2: create ad group unless already created
	if item.GAdsAdGroupId == 0 {

		// step must be 1
		if item.Step != 1 {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, fmt.Sprintf("inconsistent state: step must be 1, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Ad Group) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create ad group
		adGroupId, err := fakeGAdsApiCreateAdGroup(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Ad Group) failed: %w", err2)
			}
			return nil
		}

		// set ad group id and step to 2
		item.GAdsAdGroupId = adGroupId
		item.Step = 2
		err = svc.GAdsLaunchStore.SetAdGroupId(ctx, adGroupId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetAdGroupId failed: %w", err)
		}
	}

	// step 3: create ad unless already created
	if item.GAdsAdId == 0 {

		// step must be 2
		if item.Step != 2 {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, fmt.Sprintf("inconsistent state: step must be 2, but is %d", item.Step), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Ad) failed: %w", err2)
			}
			return nil
		}

		// call fake API to create ad
		adId, err := fakeGAdsApiCreateAd(ctx)
		if err != nil {

			// set status to Failed and write error message
			err2 := svc.GAdsLaunchStore.SetFailed(ctx, err.Error(), item.Id)
			if err2 != nil {
				return fmt.Errorf("svc.GAdsLaunchStore.SetFailed (Ad) failed: %w", err2)
			}
			return nil
		}

		// set ad id and step to 3
		item.GAdsAdId = adId
		item.Step = 3
		err = svc.GAdsLaunchStore.SetAdId(ctx, adId, item.Id)
		if err != nil {
			return fmt.Errorf("svc.GAdsLaunchStore.SetAdId failed: %w", err)
		}
	}

	// set status to Completed
	err = svc.GAdsLaunchStore.SetStatus(ctx, launchstatus.Completed, item.Id)
	if err != nil {
		return fmt.Errorf("svc.GAdsLaunchStore.SetStatus (Completed) failed: %w", err)
	}

	return nil
}
