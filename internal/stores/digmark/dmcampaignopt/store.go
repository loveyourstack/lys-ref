package dmcampaignopt

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/perfperiod"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

/*
	uses PG setFunc, not view
*/

const (
	name           string = "Digmark campaign optimizer"
	schemaName     string = "digmark"
	tableName      string = "campaign" // campaign table is base in setFunc: ok to use for rowcounts
	setFuncName    string = "campaign_optimizer"
	defaultOrderBy string = "name"
)

type Model struct {
	Id                 int64        `db:"id" json:"id"`
	Clicks             int          `db:"clicks" json:"clicks"`
	Conversions        int          `db:"conversions" json:"conversions"`
	Country            string       `db:"country" json:"country,omitempty"`
	CountryFk          int64        `db:"country_fk" json:"country_fk"`
	CountryIso2        string       `db:"country_iso2" json:"country_iso2,omitempty"`
	DailyBudgetEur     float64      `db:"daily_budget_eur" json:"daily_budget_eur"`
	EndDay             lystype.Date `db:"end_day" json:"end_day,omitzero"`
	Impressions        int          `db:"impressions" json:"impressions"`
	IsActive           bool         `db:"is_active" json:"is_active"`
	Manager            string       `db:"manager" json:"manager,omitempty"`
	Name               string       `db:"name" json:"name,omitempty"`
	ProfitEur          float64      `db:"profit_eur" json:"profit_eur"`
	ReturnOnInvestment float64      `db:"return_on_investment" json:"return_on_investment"`
	RevenueEur         float64      `db:"revenue_eur" json:"revenue_eur"`
	SpendEur           float64      `db:"spend_eur" json:"spend_eur"`
	StartDay           lystype.Date `db:"start_day" json:"start_day,omitzero"`
	Trend              float64      `db:"trend" json:"trend"`
	Vertical           string       `db:"vertical" json:"vertical,omitempty"`
	VerticalFk         int64        `db:"vertical_fk" json:"vertical_fk"`
	Volatility         float64      `db:"volatility" json:"volatility"`
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, setFuncName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}
func (s Store) GetSetFuncUrlParamNames() []string {
	return []string{"period"}
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {

	// validate period sent via mandatory API param
	period := fmt.Sprintf("%s", params.SetFuncParamValues[0])
	if !slices.Contains(perfperiod.All[:], perfperiod.Enum(period)) {
		return nil, lyspg.TotalCount{}, lyserr.User{Message: fmt.Sprintf("invalid period value: %s", period)}
	}

	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, setFuncName, defaultOrderBy, plan.DbNames(), params)
}
