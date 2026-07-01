package dmlaunchgads

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

	// GAds - account - countryIso2 - vertical

	partsS := strings.Split(name, " - ")
	if len(partsS) != 4 {
		return CampaignNameParts{}, fmt.Errorf("invalid campaign name: %s", name)
	}
	if partsS[0] != "GAds" {
		return CampaignNameParts{}, fmt.Errorf("invalid campaign name prefix: %s", partsS[0])
	}

	parts.Account = partsS[1]
	parts.CountryIso2 = partsS[2]
	parts.Vertical = partsS[3]

	return parts, nil
}
