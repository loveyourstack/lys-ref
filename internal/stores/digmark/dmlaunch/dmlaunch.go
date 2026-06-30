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
	name           string = "Digmark campaign launchers"
	schemaName     string = "digmark"
	tableName      string = "launcher"
	viewName       string = "launcher"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

// abstract table containing no records. Only selection allowed. Selection returns child table records

// Input contains the shared input fields for all launchers.
type Input struct {
	Account        string  `db:"account" json:"account,omitempty" validate:"required,max=64"`
	DailyBudgetEur float64 `db:"daily_budget_eur" json:"daily_budget_eur,omitempty" validate:"gte=0,lte=10000"`
	Manager        string  `db:"manager" json:"manager,omitempty" validate:"required,max=64"`
	Name           string  `db:"name" json:"name,omitempty" validate:"required,max=256"`
}

// Computed contains the shared computed fields for all launchers.
type Computed struct {
	CountryFk  int64             `db:"country_fk" json:"country_fk,omitempty"`   // set during preparation
	Message    string            `db:"message" json:"message,omitempty"`         // set during preparation and processing
	Status     launchstatus.Enum `db:"status" json:"status,omitempty"`           // set during preparation and processing
	VerticalFk int64             `db:"vertical_fk" json:"vertical_fk,omitempty"` // set during preparation
}

// DbManaged contains the shared database-managed fields for all launchers.
type DbManaged struct {
	Id           int64            `db:"id" json:"id,omitempty"`
	CreatedAt    lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CreatedAtDay lystype.Date     `db:"created_at_day" json:"created_at_day,omitzero"`
	Partner      string           `db:"partner" json:"partner,omitempty"`
	UpdatedAt    lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
}

type Model struct {
	Input
	Computed
	DbManaged
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
