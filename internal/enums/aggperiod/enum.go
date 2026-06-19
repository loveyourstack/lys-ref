package aggperiod

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Day   Enum = "day"
	Week  Enum = "week"
	Month Enum = "month"
	Year  Enum = "year"
)

var (
	All     = [...]Enum{Day, Week, Month, Year}
	Default = Week
)

func DaysBefore(period Enum) int {
	switch period {
	case Day:
		return 1
	case Week:
		return 7
	case Month:
		return 30
	case Year:
		return 365
	}
	return 7
}

func FromString(s string) Enum {
	for _, p := range All {
		if p.String() == s {
			return p
		}
	}
	return Default
}
