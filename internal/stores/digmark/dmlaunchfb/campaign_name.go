package dmlaunchfb

import (
	"fmt"
	"strings"
)

type CampaignNameParts struct {
	Account     string
	CountryIso2 string
	Vertical    string
}

func ParseCampaignName(name string) (parts CampaignNameParts, err error) {

	// account - countryIso2 - vertical

	partsS := strings.Split(name, " - ")
	if len(partsS) != 3 {
		return CampaignNameParts{}, fmt.Errorf("invalid campaign name: %s", name)
	}

	parts.Account = partsS[0]
	parts.CountryIso2 = partsS[1]
	parts.Vertical = partsS[2]

	return parts, nil
}
