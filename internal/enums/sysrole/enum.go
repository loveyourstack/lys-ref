// sysrole represents enum system.role
package sysrole

import (
	"fmt"
	"slices"
)

type Enum string

func (e Enum) String() string {
	return string(e)
}

func FromStringSlice(s []string) ([]Enum, error) {
	enums := make([]Enum, len(s))
	for i, str := range s {
		if !slices.Contains(All, Enum(str)) {
			return nil, fmt.Errorf("invalid sysrole: %s", str)
		}
		enums[i] = Enum(str)
	}
	return enums, nil
}

func ToStringSlice(enums []Enum) []string {
	strs := make([]string, len(enums))
	for i, e := range enums {
		strs[i] = e.String()
	}
	return strs
}

const (
	Standard Enum = "Standard"
	Tech     Enum = "Tech"
	Viewer   Enum = "Viewer"
)

var (
	All    = []Enum{Standard, Tech, Viewer}
	Writer = [...]Enum{Standard, Tech}
)
