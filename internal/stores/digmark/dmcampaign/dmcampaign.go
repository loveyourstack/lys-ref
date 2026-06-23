package dmcampaign

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Digmark campaigns"
	schemaName     string = "digmark"
	tableName      string = "campaign"
	viewName       string = "v_campaign"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

type Input struct {
	CountryFk      int64   `db:"country_fk" json:"country_fk,omitempty" validate:"required,gte=1"` // disallow -1 (None)
	DailyBudgetEur float64 `db:"daily_budget_eur" json:"daily_budget_eur" validate:"gte=0,lte=10000"`
	IsActive       bool    `db:"is_active" json:"is_active,omitempty"`
	Manager        string  `db:"manager" json:"manager,omitempty"`
	Name           string  `db:"name" json:"name,omitempty" validate:"max=256"`
	VerticalFk     int64   `db:"vertical_fk" json:"vertical_fk,omitempty" validate:"required"`
}

type Model struct {
	Id               int64            `db:"id" json:"id,omitempty"`
	Country          string           `db:"country" json:"country,omitempty"`
	CountryIso2      string           `db:"country_iso2" json:"country_iso2,omitempty"`
	CreatedAt        lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	PerformanceRange string           `db:"performance_range" json:"performance_range,omitempty"`
	Vertical         string           `db:"vertical" json:"vertical,omitempty"`
	UpdatedAt        lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
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

func (s Store) Delete(ctx context.Context, id int64) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Insert(ctx context.Context, input Input) (newId int64, err error) {
	return lyspg.Insert[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SelectIdBudgetMap(ctx context.Context, ids []int64) (idBudgetMap map[int64]float64, err error) {

	stmt := fmt.Sprintf("SELECT id, daily_budget_eur FROM %s.%s WHERE id = ANY($1)", schemaName, viewName)
	camps, err := lyspg.SelectT[Model](ctx, s.Db, stmt, ids)
	if err != nil {
		return nil, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	idBudgetMap = make(map[int64]float64)
	for _, camp := range camps {
		idBudgetMap[camp.Id] = camp.DailyBudgetEur
	}

	return idBudgetMap, nil
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func UpdatePartialTx(ctx context.Context, tx pgx.Tx, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, tx, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
