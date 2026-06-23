// perfperiod represents enum core.performance_period
package perfperiod

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Today      Enum = "Today"
	Yesterday  Enum = "Yesterday"
	Last3Days  Enum = "Last 3 days"
	Last7Days  Enum = "Last 7 days"
	Last14Days Enum = "Last 14 days"
	Last30Days Enum = "Last 30 days"
	ThisMonth  Enum = "This month"
	LastMonth  Enum = "Last month"
	Last90Days Enum = "Last 90 days"
)

var (
	All = [...]Enum{Today, Yesterday, Last3Days, Last7Days, Last14Days, Last30Days, ThisMonth, LastMonth, Last90Days}
	Xr  = [...]Enum{Last7Days, Last14Days, Last30Days, ThisMonth, LastMonth, Last90Days}
)
