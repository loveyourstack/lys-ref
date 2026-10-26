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
	Monitor  Enum = "Monitor"  // for gRPC server service account: allows access to internal auth structs
	Standard Enum = "Standard" // standard user read and write
	Tech     Enum = "Tech"     // technical super user
	Viewer   Enum = "Viewer"   // read-only user
)

var (
	All    = []Enum{Monitor, Standard, Tech, Viewer}
	Writer = [...]Enum{Standard, Tech}
)
