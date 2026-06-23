package corevarianttype

import (
	"context"
	"log"
	"net/netip"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Variant types"
	schemaName     string = "core"
	tableName      string = "variant_type"
	viewName       string = "v_variant_type"
	pkColName      string = "id"
	defaultOrderBy string = "id DESC"
)

type Input struct {
	CConstrainedText string     `db:"c_constrained_text" json:"c_constrained_text,omitempty" validate:"required,len=6,uppercase"`
	CIp              netip.Addr `db:"c_ip" json:"c_ip,omitzero" validate:"required"` // don't use "ip" in validate: it requires subnet mask
	CLongText        string     `db:"c_long_text" json:"c_long_text,omitempty" validate:"required,max=1000"`
	CMoneyAmount     float64    `db:"c_money_amount" json:"c_money_amount,omitempty"`
	CPercent         float64    `db:"c_percent" json:"c_percent,omitempty" validate:"min=0,max=10"`
}

type Model struct {
	Id             int64            `db:"id" json:"id,omitempty"`
	CLongTextShort string           `db:"c_long_text_short" json:"c_long_text_short,omitempty"`
	CreatedAt      lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	UpdatedAt      lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
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

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
