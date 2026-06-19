package dmcampperf

import (
	"context"
	"fmt"
	"strings"

	"github.com/loveyourstack/lys-ref/internal/enums/aggperiod"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"github.com/loveyourstack/lys/lystype"
)

type Metrics struct {
	Clicks             int     `db:"clicks" json:"clicks"`
	Conversions        int     `db:"conversions" json:"conversions"`
	Impressions        int     `db:"impressions" json:"impressions"`
	ProfitEur          float64 `db:"profit_eur" json:"profit_eur"`
	ReturnOnInvestment float64 `db:"return_on_investment" json:"return_on_investment"`
	RevenueEur         float64 `db:"revenue_eur" json:"revenue_eur"`
	SpendEur           float64 `db:"spend_eur" json:"spend_eur"`
	Vertical           string  `db:"vertical" json:"vertical"`
}

var metricDbCols = lysset.New("clicks", "conversions", "impressions", "profit_eur", "return_on_investment", "revenue_eur", "spend_eur")

type ReportingWindow struct {
	EndDay   lystype.Date `db:"end_day" json:"end_day"`
	StartDay lystype.Date `db:"start_day" json:"start_day"`
}

type DailyTrend struct {
	Metrics
	Day lystype.Date `db:"day" json:"day"`
}

// SelectDailyTrend returns per-day aggregated performance for the past N days.
func (s Store) SelectDailyTrend(ctx context.Context, days int) ([]DailyTrend, error) {

	if days <= 0 || days > 31 {
		days = 7
	}

	stmt := fmt.Sprintf(`SELECT %s, day FROM %s.daily_trend($1) ORDER BY day;`, strings.Join(metricDbCols.Values(), ", "), schemaName)
	return lyspg.SelectT[DailyTrend](ctx, s.Db, stmt, days)
}

type LatestPerfSummary struct {
	Day          lystype.Date `db:"day" json:"day"`
	TotalRevenue float64      `db:"total_revenue" json:"total_revenue"`
	TotalSpend   float64      `db:"total_spend" json:"total_spend"`
}

// SelectLatestPerfSummary returns a summary of daily revenue and spend performance for the most recent week.
func (s Store) SelectLatestPerfSummary(ctx context.Context) (items []LatestPerfSummary, err error) {
	stmt := fmt.Sprintf(`SELECT day, total_spend, total_revenue FROM %s.v_latest_perf_summary;`, schemaName)
	return lyspg.SelectT[LatestPerfSummary](ctx, s.Db, stmt)
}

type PerfSummary struct {
	Metrics
	ReportingWindow
	ActiveCampaigns int `db:"active_campaigns" json:"active_campaigns"`
}

// SelectPerfSummary returns a single-row aggregate of all metrics for the given period.
func (s Store) SelectPerfSummary(ctx context.Context, period aggperiod.Enum) (PerfSummary, error) {
	daysBack := aggperiod.DaysBefore(period)

	stmt := fmt.Sprintf(`SELECT %s, active_campaigns, end_day, start_day FROM %s.perf_summary($1);`, strings.Join(metricDbCols.Values(), ", "), schemaName)
	rows, err := lyspg.SelectT[PerfSummary](ctx, s.Db, stmt, daysBack)
	if err != nil {
		return PerfSummary{}, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}
	if len(rows) == 0 {
		return PerfSummary{}, fmt.Errorf("no data found for period %v", period)
	}
	return rows[0], nil
}

type VerticalPerf struct {
	Metrics
	ReportingWindow
	Vertical string `db:"vertical" json:"vertical"`
}

// SelectVerticalPerformance returns campaign performance aggregated by vertical for the given period.
// orderBy is one of the metricDbCols, defaulting to profit_eur if not recognized.
func (s Store) SelectVerticalPerformance(ctx context.Context, period aggperiod.Enum, orderBy string) ([]VerticalPerf, error) {
	daysBack := aggperiod.DaysBefore(period)

	// orderBy defaults to profit, but must be a metric
	orderCol := "profit_eur"
	if metricDbCols.Contains(orderBy) {
		orderCol = orderBy
	}

	stmt := fmt.Sprintf(`SELECT %s, end_day, start_day, vertical
		FROM %s.vertical_perf($1) ORDER BY %s DESC;`, strings.Join(metricDbCols.Values(), ", "), schemaName, orderCol)

	return lyspg.SelectT[VerticalPerf](ctx, s.Db, stmt, daysBack)
}

type TopCampaigns struct {
	Metrics
	ReportingWindow
	Campaign string `db:"campaign" json:"campaign"`
	Manager  string `db:"manager" json:"manager"`
	Vertical string `db:"vertical" json:"vertical"`
}

// SelectTopCampaigns returns the top N campaigns ranked by metric for the given period.
// orderBy is one of the metricDbCols, defaulting to profit_eur if not recognized.
func (s Store) SelectTopCampaigns(ctx context.Context, period aggperiod.Enum, orderBy string, limit int) ([]TopCampaigns, error) {
	daysBack := aggperiod.DaysBefore(period)

	// orderBy defaults to profit, but must be a metric
	orderCol := "profit_eur"
	if metricDbCols.Contains(orderBy) {
		orderCol = orderBy
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	stmt := fmt.Sprintf(`SELECT %s, campaign, end_day, start_day, manager, vertical
		FROM %s.top_campaigns($1) ORDER BY %s DESC LIMIT $2;`, strings.Join(metricDbCols.Values(), ", "), schemaName, orderCol)

	return lyspg.SelectT[TopCampaigns](ctx, s.Db, stmt, daysBack, limit)
}
