package dmcampaignopt

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/internal/enums/perfperiod"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

type AggModel struct {
	Clicks             int     `db:"clicks" json:"clicks"`
	Conversions        int     `db:"conversions" json:"conversions"`
	DailyBudgetEur     float64 `db:"daily_budget_eur" json:"daily_budget_eur"`
	Impressions        int     `db:"impressions" json:"impressions"`
	IsActive           int     `db:"is_active" json:"is_active"`
	ProfitEur          float64 `db:"profit_eur" json:"profit_eur"`
	ReturnOnInvestment float32 `db:"return_on_investment" json:"return_on_investment"`
	RevenueEur         float64 `db:"revenue_eur" json:"revenue_eur"`
	SpendEur           float64 `db:"spend_eur" json:"spend_eur"`
}

func (s Store) SelectAggregates(ctx context.Context, setFuncParamValues []any, conditions []lyspg.Condition) (item AggModel, err error) {

	// validate period
	period := fmt.Sprintf("%s", setFuncParamValues[0])
	if !slices.Contains(perfperiod.All[:], perfperiod.Enum(period)) {
		return AggModel{}, lyserr.User{Message: fmt.Sprintf("invalid period value: %s", period)}
	}

	sourceName := lyspg.GetSourceName(setFuncName, len(setFuncParamValues))
	whereClause, _ := lyspg.GetWhereClause(len(setFuncParamValues), conditions, nil)

	stmt := fmt.Sprintf(`SELECT 
		COALESCE(SUM(clicks),0) AS clicks,
		COALESCE(SUM(conversions),0) AS conversions,
		COALESCE(SUM(daily_budget_eur),0) AS daily_budget_eur,
		COALESCE(SUM(impressions),0) AS impressions,
		COUNT(*) FILTER (WHERE is_active = true) AS is_active,
		COALESCE(SUM(revenue_eur) - SUM(spend_eur),0) AS profit_eur,
		CASE WHEN SUM(spend_eur) = 0 THEN 0.0 ELSE COALESCE(SUM(revenue_eur)/SUM(spend_eur) - 1,0) END AS return_on_investment,
		COALESCE(SUM(revenue_eur),0) AS revenue_eur,
		COALESCE(SUM(spend_eur),0) AS spend_eur
	FROM %s.%s WHERE 1=1%s`, schemaName, sourceName, whereClause)

	paramValues := lyspg.GetSelectParamValues(setFuncParamValues, conditions, nil, false, 0, 0)

	rows, _ := s.Db.Query(ctx, stmt, paramValues...)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[AggModel])
	if err != nil {
		return item, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return item, nil
}
