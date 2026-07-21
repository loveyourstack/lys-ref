package tedbvatratesumm

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "TEDB VAT rate summary"
	schemaName     string = "tedb"
	tableName      string = "v_vat_rate_summary"
	viewName       string = "v_vat_rate_summary"
	defaultOrderBy string = "country, type DESC"
)

type Model struct {
	Categories  string       `db:"categories" json:"categories,omitempty"`
	Comment     string       `db:"comment" json:"comment,omitempty"`
	Country     string       `db:"country" json:"country,omitempty"`
	Rate        float64      `db:"rate" json:"rate"`
	SituationOn lystype.Date `db:"situation_on" json:"situation_on,omitzero"`
	Type        string       `db:"type" json:"type,omitempty"`
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}
