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

type PrepareFbHelpers struct {
	fbAccountIdMap map[string]string
	prepareHelpers
}

func (svc Service) selectPrepareFbHelpers(ctx context.Context) (helpers PrepareFbHelpers, err error) {
	helpers.prepareHelpers, err = svc.selectPrepareHelpers(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.selectPrepareHelpers failed: %w", err)
	}

	// fake fb account id map: would normally be selected from store
	helpers.fbAccountIdMap = map[string]string{
		"FB 1": "ABC",
		"FB 2": "DEF",
	}

	return helpers, nil
}

type PrepareFbParams struct {
	Account string
	prepareParams
}

type PrepareFbResult struct {
	FbAccountId string
	prepareResult
}

func PrepareFb(helpers PrepareFbHelpers, params PrepareFbParams) (result PrepareFbResult, err error) {
	result.prepareResult, err = prepare(helpers.prepareHelpers, params.prepareParams)

	if fbAccountId, ok := helpers.fbAccountIdMap[params.Account]; ok {
		result.FbAccountId = fbAccountId
	} else {
		err = errors.Join(err, fmt.Errorf("invalid account: %s", params.Account))
	}

	return result, err
}

// -------------------------------------------------------------------------------------------------------------------

type PrepareGAdsHelpers struct {
	gadsAccountIdMap map[string]int64
	prepareHelpers
}

func (svc Service) selectPrepareGAdsHelpers(ctx context.Context) (helpers PrepareGAdsHelpers, err error) {
	helpers.prepareHelpers, err = svc.selectPrepareHelpers(ctx)
	if err != nil {
		return helpers, fmt.Errorf("svc.selectPrepareHelpers failed: %w", err)
	}

	// fake gads account id map: would normally be selected from store
	helpers.gadsAccountIdMap = map[string]int64{
		"GAds 1": 123,
		"GAds 2": 456,
	}

	return helpers, nil
}

type PrepareGAdsParams struct {
	Account string
	prepareParams
}

type PrepareGAdsResult struct {
	GAdsAccountId int64
	prepareResult
}

func PrepareGAds(helpers PrepareGAdsHelpers, params PrepareGAdsParams) (result PrepareGAdsResult, err error) {
	result.prepareResult, err = prepare(helpers.prepareHelpers, params.prepareParams)

	if gadsAccountId, ok := helpers.gadsAccountIdMap[params.Account]; ok {
		result.GAdsAccountId = gadsAccountId
	} else {
		err = errors.Join(err, fmt.Errorf("invalid account: %s", params.Account))
	}

	return result, err
}
