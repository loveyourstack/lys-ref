package launchsvc

import (
	"context"
	"errors"
	"fmt"
)

type prepareHelpers struct {
	CountryIso2IdMap  map[string]int64
	VerticalNameIdMap map[string]int64
}

func (svc Service) selectPrepareHelpers(ctx context.Context) (helpers prepareHelpers, err error) {

	helpers.CountryIso2IdMap, err = svc.CoStore.Iso2IdValueMap(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.CoStore.Iso2IdValueMap failed: %w", err)
	}

	helpers.VerticalNameIdMap, err = svc.VertStore.NameIdValueMap(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.VertStore.NameIdValueMap failed: %w", err)
	}

	return helpers, nil
}

type prepareParams struct {
	CountryIso2 string
	Vertical    string
}

type prepareResult struct {
	CountryFk  int64
	VerticalFk int64
}

func prepare(helpers prepareHelpers, params prepareParams) (result prepareResult, err error) {

	if countryFk, ok := helpers.CountryIso2IdMap[params.CountryIso2]; ok {
		result.CountryFk = countryFk
	} else {
		err = errors.Join(err, fmt.Errorf("invalid country: %s", params.CountryIso2))
	}

	if verticalFk, ok := helpers.VerticalNameIdMap[params.Vertical]; ok {
		result.VerticalFk = verticalFk
	} else {
		err = errors.Join(err, fmt.Errorf("invalid vertical: %s", params.Vertical))
	}

	return result, err
}

// -------------------------------------------------------------------------------------------------------------------

type prepareFbHelpers struct {
	fbAccountIdMap map[string]string
	prepareHelpers
}

func (svc Service) selectPrepareFbHelpers(ctx context.Context) (helpers prepareFbHelpers, err error) {
	helpers.prepareHelpers, err = svc.selectPrepareHelpers(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.selectPrepareHelpers failed: %w", err)
	}

	// fake fb account id map: would normally be selected from store
	helpers.fbAccountIdMap = map[string]string{
		"Acc1": "ABC",
		"Acc2": "DEF",
	}

	return helpers, nil
}

type prepareFbParams struct {
	Account string
	prepareParams
}

type prepareFbResult struct {
	FbAccountId string
	prepareResult
}

func prepareFb(helpers prepareFbHelpers, params prepareFbParams) (result prepareFbResult, err error) {
	result.prepareResult, err = prepare(helpers.prepareHelpers, params.prepareParams)

	if fbAccountId, ok := helpers.fbAccountIdMap[params.Account]; ok {
		result.FbAccountId = fbAccountId
	} else {
		err = errors.Join(err, fmt.Errorf("invalid account: %s", params.Account))
	}

	return result, err
}

// -------------------------------------------------------------------------------------------------------------------

type prepareGAdsHelpers struct {
	gadsAccountIdMap map[string]int64
	prepareHelpers
}

func (svc Service) selectPrepareGAdsHelpers(ctx context.Context) (helpers prepareGAdsHelpers, err error) {
	helpers.prepareHelpers, err = svc.selectPrepareHelpers(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.selectPrepareHelpers failed: %w", err)
	}

	// fake gads account id map: would normally be selected from store
	helpers.gadsAccountIdMap = map[string]int64{
		"Acc1": 123,
		"Acc2": 456,
	}

	return helpers, nil
}

type prepareGAdsParams struct {
	Account string
	prepareParams
}

type prepareGAdsResult struct {
	GAdsAccountId int64
	prepareResult
}

func prepareGAds(helpers prepareGAdsHelpers, params prepareGAdsParams) (result prepareGAdsResult, err error) {
	result.prepareResult, err = prepare(helpers.prepareHelpers, params.prepareParams)

	if gadsAccountId, ok := helpers.gadsAccountIdMap[params.Account]; ok {
		result.GAdsAccountId = gadsAccountId
	} else {
		err = errors.Join(err, fmt.Errorf("invalid account: %s", params.Account))
	}

	return result, err
}
