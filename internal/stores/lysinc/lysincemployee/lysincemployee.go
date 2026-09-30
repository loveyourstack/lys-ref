package lysincemployee

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystree"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Lys Inc employees"
	schemaName     string = "lysinc"
	tableName      string = "employee"
	viewName       string = "v_employee"
	pkColName      string = "id"
	defaultOrderBy string = "full_name"
)

type Input struct {
	DateOfBirth  lystype.Date `db:"date_of_birth" json:"date_of_birth,omitzero" validate:"required"`
	DepartmentFk int64        `db:"department_fk" json:"department_fk,omitzero" validate:"required"`
	Email        string       `db:"email" json:"email,omitzero" validate:"required,email,max=256"`
	FamilyName   string       `db:"family_name" json:"family_name,omitzero" validate:"required,max=256"`
	GivenName    string       `db:"given_name" json:"given_name,omitzero" validate:"required,max=256"`
	Honorific    string       `db:"honorific" json:"honorific,omitzero" validate:"max=64,required"`
	JobTitle     string       `db:"job_title" json:"job_title,omitzero" validate:"required,max=256"`
	JoinDate     lystype.Date `db:"join_date" json:"join_date,omitzero" validate:"required"`
	ProfilePic   string       `db:"profile_pic" json:"profile_pic,omitzero" validate:"max=256,required"`
	ReportsTo    int64        `db:"reports_to" json:"reports_to,omitzero" validate:"required"`
	Sex          string       `db:"sex" json:"sex,omitzero" validate:"required"`
}

type Model struct {
	Id                int64            `db:"id" json:"id,omitzero"`
	CreatedAt         lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	Department        string           `db:"department" json:"department,omitzero"`
	FullName          string           `db:"full_name" json:"full_name,omitzero"`
	ReportsToFullName string           `db:"reports_to_full_name" json:"reports_to_full_name,omitzero"`
	ReportsToJobTitle string           `db:"reports_to_job_title" json:"reports_to_job_title,omitzero"`
	UpdatedAt         lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
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

// SelectTree returns all employees in a hierarchical tree structure based on ReportsTo.
func (s Store) SelectTree(ctx context.Context) (nodes []*lystree.Node[Model], err error) {

	items, _, err := s.Select(ctx, lyspg.SelectParams{
		Fields: []string{ // only select the fields needed for UI tree view
			"department",
			"full_name",
			"id",
			"job_title",
			"profile_pic",
			"reports_to",
			"reports_to_full_name",
			"reports_to_job_title",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("s.Select failed: %w", err)
	}

	return lystree.FromItems(
		items,
		func(item Model) int64 { return item.Id },
		func(item Model) int64 { return item.ReportsTo },
	)
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
