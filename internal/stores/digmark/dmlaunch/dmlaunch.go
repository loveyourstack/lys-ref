package dmlaunch

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Digmark launchers"
	schemaName     string = "digmark"
	tableName      string = "launcher"
	viewName       string = "launcher"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

// abstract table - only selection allowed

type Input struct {
	DailyBudgetEur float64           `db:"daily_budget_eur" json:"daily_budget_eur,omitempty" validate:"gte=0,lte=10000"`
	Manager        string            `db:"manager" json:"manager,omitempty" validate:"max=64"`
	Message        string            `db:"message" json:"message,omitempty"`
	Name           string            `db:"name" json:"name,omitempty" validate:"max=256"`
	Partner        string            `db:"partner" json:"partner,omitempty" validate:"max=64"`
	Status         launchstatus.Enum `db:"status" json:"status,omitempty" validate:"max=64"`
}

type Model struct {
	Id           int64            `db:"id" json:"id,omitempty"`
	CreatedAt    lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CreatedAtDay lystype.Date     `db:"created_at_day" json:"created_at_day,omitzero"`
	UpdatedAt    lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
	Input
}

var (
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.Analyze(Input{})
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

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}
