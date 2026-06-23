package procrun

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Process run"
	schemaName     string = "process"
	tableName      string = "run"
	viewName       string = "v_run"
	pkColName      string = "id"
	defaultOrderBy string = "id DESC"
)

type Input struct {
	FlowFk   int64  `db:"flow_fk" json:"flow_fk,omitempty" validate:"required"`
	StepId   int64  `db:"step_id" json:"step_id,omitempty" validate:"required"`
	StepName string `db:"step_name" json:"step_name,omitempty" validate:"required"`
}

type Model struct {
	Id         int64            `db:"id" json:"id"`
	CreatedAt  lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	FinishedAt lystype.Datetime `db:"finished_at" json:"finished_at,omitzero"`
	Flow       string           `db:"flow" json:"flow,omitempty"`
	PointCount int              `db:"point_count" json:"point_count,omitempty"`
	PointStati string           `db:"point_stati" json:"point_stati,omitempty"`
	StartedAt  lystype.Datetime `db:"started_at" json:"started_at,omitzero"`
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

func InsertTx(ctx context.Context, tx pgx.Tx, input Input) (newId int64, err error) {
	return lyspg.Insert[Input, int64](ctx, tx, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
