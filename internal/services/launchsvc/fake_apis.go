package launchsvc

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysstring"
)

func fakeFbApiCall(ctx context.Context) (fakeId string, err error) {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// simulate an API delay but allow context cancellation
	timer := time.NewTimer(time.Duration(2+rng.Intn(5)) * time.Second)
	select {
	case <-ctx.Done():
		timer.Stop()
		return "", ctx.Err()
	case <-timer.C:
	}

	switch rng.Intn(10) {
	case 0:
		return "", lyserr.Ext{Err: fmt.Errorf("Facebook API error"), Message: "Facebook API error"}
	default:
		return lysstring.Rand(8), nil
	}
}

func fakeFbApiCreateCreative(ctx context.Context) (creativeId string, err error) {
	return fakeFbApiCall(ctx)
}
func fakeFbApiCreateCampaign(ctx context.Context) (campaignId string, err error) {
	return fakeFbApiCall(ctx)
}

func fakeGAdsApiCall(ctx context.Context) (fakeId int64, err error) {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// simulate an API delay but allow context cancellation
	timer := time.NewTimer(time.Duration(2+rng.Intn(5)) * time.Second)
	select {
	case <-ctx.Done():
		timer.Stop()
		return 0, ctx.Err()
	case <-timer.C:
	}

	switch rng.Intn(10) {
	case 0:
		return 0, lyserr.Ext{Err: fmt.Errorf("Google Ads API error"), Message: "Google Ads API error"}
	default:
		return int64(1000000 + rng.Intn(1000000)), nil
	}
}

func fakeGAdsApiCreateAd(ctx context.Context) (adId int64, err error) {
	return fakeGAdsApiCall(ctx)
}
func fakeGAdsApiCreateAdGroup(ctx context.Context) (adGroupId int64, err error) {
	return fakeGAdsApiCall(ctx)
}
func fakeGAdsApiCreateCampaign(ctx context.Context) (campaignId int64, err error) {
	return fakeGAdsApiCall(ctx)
}
