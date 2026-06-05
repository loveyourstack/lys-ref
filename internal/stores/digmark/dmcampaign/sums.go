package dmcampaign

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys/lyspg"
)

type managerBudget struct {
	Manager     string `db:"manager" json:"manager"`
	TotalBudget int    `db:"total_budget" json:"total_budget,omitempty"`
}

func (s Store) SelectManagerBudgets(ctx context.Context) (items []managerBudget, err error) {
	stmt := fmt.Sprintf(`SELECT manager, SUM(daily_budget_eur)::int AS total_budget 
		FROM %s.%s 
		WHERE is_active = true
		GROUP BY 1 ORDER BY 2 DESC;`, schemaName, tableName)
	return lyspg.SelectT[managerBudget](ctx, s.Db, stmt)
}

type verticalBudget struct {
	Vertical    string `db:"vertical" json:"vertical"`
	TotalBudget int    `db:"total_budget" json:"total_budget,omitempty"`
}

func (s Store) SelectVerticalBudgets(ctx context.Context) (items []verticalBudget, err error) {
	stmt := fmt.Sprintf(`SELECT dm_v.name AS vertical, SUM(dm_c.daily_budget_eur)::int AS total_budget 
		FROM %s.%s dm_c JOIN digmark.vertical dm_v ON dm_c.vertical_fk = dm_v.id
		WHERE dm_c.is_active = true 
		GROUP BY 1 ORDER BY 2 DESC;`, schemaName, tableName)
	return lyspg.SelectT[verticalBudget](ctx, s.Db, stmt)
}
