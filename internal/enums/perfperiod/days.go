package perfperiod

import (
	"fmt"
	"time"

	"github.com/loveyourstack/lys/lyserr"
)

func startOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

/*func endOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}*/

func Days(period Enum, baseTime time.Time) (daysBefore, daysAfter int, err error) {

	switch period {
	case Today:
		// both 0
	case Yesterday:
		daysBefore = -1
		daysAfter = -1
	case Last3Days: // including base day
		daysBefore = -2
	case Last7Days: // including base day
		daysBefore = -6
	case Last14Days: // including base day
		daysBefore = -13
	case Last30Days: // including base day
		daysBefore = -29
	case ThisMonth: // including base day
		start := startOfMonth(baseTime)
		daysBefore = -int(baseTime.Sub(start).Hours() / 24)
	case LastMonth:
		thisMonthStart := startOfMonth(baseTime)
		lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
		daysBefore = -int(baseTime.Sub(lastMonthStart).Hours() / 24)
		daysAfter = -int(baseTime.Sub(thisMonthStart).Hours()/24) - 1
	case Last90Days: // including base day
		daysBefore = -89
	default:
		return 0, 0, lyserr.User{Message: fmt.Sprintf("unknown period: %s", period)}
	}

	return daysBefore, daysAfter, nil
}
