package dmcampperf

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

type latestPerfSummary struct {
	Day          lystype.Date `db:"day" json:"day"`
	TotalSpend   float64      `db:"total_spend" json:"total_spend"`
	TotalRevenue float64      `db:"total_revenue" json:"total_revenue"`
}

func (s Store) SelectLatestPerfSummary(ctx context.Context) (items []latestPerfSummary, err error) {
	stmt := fmt.Sprintf(`SELECT day_cet AS day, SUM(spend_eur) AS total_spend, SUM(revenue_eur) AS total_revenue 
		FROM %s.%s 
		WHERE day_cet > current_date -7
		GROUP BY 1 ORDER BY 1`, schemaName, tableName)
	return lyspg.SelectT[latestPerfSummary](ctx, s.Db, stmt)
}
