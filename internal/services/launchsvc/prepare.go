package launchsvc

import (
	"errors"
	"fmt"
)

type PrepareParams struct {
	CountryIso2IdMap  map[string]int64
	VerticalNameIdMap map[string]int64
}

type PrepareResult struct {
	CountryFk  int64
	VerticalFk int64
}

func (svc Service) Prepare(params PrepareParams, countryIso2 string, verticalName string) (result PrepareResult, err error) {

	if countryFk, ok := params.CountryIso2IdMap[countryIso2]; ok {
		result.CountryFk = countryFk
	} else {
		err = errors.Join(err, fmt.Errorf("invalid country ISO2: %s", countryIso2))
	}

	if verticalFk, ok := params.VerticalNameIdMap[verticalName]; ok {
		result.VerticalFk = verticalFk
	} else {
		err = errors.Join(err, fmt.Errorf("invalid vertical name: %s", verticalName))
	}

	return result, err
}
