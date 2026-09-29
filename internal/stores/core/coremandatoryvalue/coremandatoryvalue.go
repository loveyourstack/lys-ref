package coremandatoryvalue

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
	name           string = "Mandatory values"
	schemaName     string = "core"
	tableName      string = "mandatory_value"
	viewName       string = "v_mandatory_value"
	pkColName      string = "id"
	defaultOrderBy string = "id DESC"
)

type Input struct {

	/*
	  json/v2 - changes for bool and numeric types

	  omitzero: When marshaling, the "omitzero" option specifies that the struct field should be omitted if the field value is zero as determined by the "IsZero() bool" method if present,
	  otherwise based on whether the field is the zero Go value

	  omitempty: When marshaling, the "omitempty" option specifies that the struct field should be omitted if the field value would have been encoded as a JSON null, empty string,
	  empty object, or empty array

	  use omitzero rather than omitempty for bool and numeric types (https://pkg.go.dev/encoding/json/v2#hdr-JSON_Representation_of_Go_structs),
	  but only when it's ok for false/0 to be omitted in json response
	  this makes xfields param work correctly (bool/numeric fields will be omitted if zero), but needs isNaN checks in UI for numeric types

	  omitempty now has no effect for bool and numeric types
	  since omitzero also works for strings ("") and pointers (nil), and was already used for custom types with an IsZero method,
	  use omitzero for all non-slice type fields where the zero value can be safely omitted.
	  Slice types should still use omitempty.
	*/

	// don't use validate:"required": it will reject false
	CBool bool `db:"c_bool" json:"c_bool,omitzero"`

	// note use of omitzero rather than omitempty for lystype types
	CDateCet lystype.Date `db:"c_date_cet" json:"c_date_cet,omitzero" validate:"required"`

	CEnum string `db:"c_enum" json:"c_enum,omitzero" validate:"required"`

	// numbers: if zero is allowed, don't use validate:"required": it will reject 0
	// use gte/lte rather than min/max to ensure correct validation message
	CInt     int64   `db:"c_int" json:"c_int,omitzero" validate:"lte=1000000"`
	CNumeric float64 `db:"c_numeric" json:"c_numeric,omitzero" validate:"lte=1000000"`

	CTableFk int64        `db:"c_table_fk" json:"c_table_fk,omitzero" validate:"required"`
	CText    string       `db:"c_text" json:"c_text,omitzero" validate:"required,max=256"`
	CTime    lystype.Time `db:"c_time" json:"c_time,omitzero"`
}

type Model struct {
	Id        int64            `db:"id" json:"id,omitzero"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CTable    string           `db:"c_table" json:"c_table,omitzero"`
	UpdatedAt lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
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

// needed for import route
func (s Store) InsertTx(ctx context.Context, tx pgx.Tx, input Input) (newId int64, err error) {
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
