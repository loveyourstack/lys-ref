package sysuser

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "System users"
	schemaName     string = "system"
	tableName      string = "user"
	viewName       string = "user"
	pkColName      string = "id"
	defaultOrderBy string = "full_name"
)

type Input struct {
	AllowMultipleSessions bool           `db:"allow_multiple_sessions" json:"allow_multiple_sessions,omitempty"`
	Email                 string         `db:"email" json:"email,omitempty" validate:"required,email,max=256"`
	FamilyName            string         `db:"family_name" json:"family_name,omitempty" validate:"required,max=256"`
	GivenName             string         `db:"given_name" json:"given_name,omitempty" validate:"required,max=256"`
	HashedPw              string         `db:"hashed_pw" json:"-"`
	Name                  string         `db:"name" json:"name,omitempty" validate:"required,max=64"`
	Roles                 []sysrole.Enum `db:"roles" json:"roles,omitempty" validate:"required"`
}

type Model struct {
	Id                  int64            `db:"id" json:"id,omitempty"`
	CreatedAt           lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CreatedBy           string           `db:"created_by" json:"created_by,omitempty"`         // assigned in Insert func
	DeactivatedAt       lystype.Datetime `db:"deactivated_at" json:"deactivated_at,omitzero"`  // set in Deactivate func
	EmailVerified       bool             `db:"email_verified" json:"email_verified,omitempty"` // set in VerifyEmail func
	ForcePasswordChange bool             `db:"force_password_change" json:"force_password_change,omitempty"`
	FullName            string           `db:"full_name" json:"full_name,omitempty"`
	IsDeactivated       bool             `db:"is_deactivated" json:"is_deactivated,omitempty"`           // set in Deactivate func
	LastUserUpdateBy    string           `db:"last_user_update_by" json:"last_user_update_by,omitempty"` // assigned in Update funcs
	UpdatedAt           lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"`                    // assigned by trigger
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

func (s Store) Deactivate(ctx context.Context, id int64) error {

	item, err := s.SelectById(ctx, id)
	if err != nil {
		return fmt.Errorf("s.SelectById failed: %w", err)
	}

	assignmentsMap := make(map[string]any)
	assignmentsMap["deactivated_at"] = "now()"
	assignmentsMap["is_deactivated"] = true
	assignmentsMap["name"] = item.Name + "_deactivated"

	err = s.UpdatePartial(ctx, assignmentsMap, id)
	if err != nil {
		return fmt.Errorf("s.UpdatePartial failed: %w", err)
	}

	return nil
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
	return lyspg.InsertWithExtras[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input,
		[]string{"created_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectIdNameMap(ctx context.Context) (idNameMap map[int64]string, err error) {
	items, _, err := s.Select(ctx, lyspg.SelectParams{})
	if err != nil {
		return nil, fmt.Errorf("s.Select failed: %w", err)
	}

	idNameMap = make(map[int64]string, len(items))
	for _, item := range items {
		idNameMap[item.Id] = item.Name
	}
	return idNameMap, nil
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SelectByName(ctx context.Context, name string) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, "name", name)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.UpdateWithExtras(ctx, s.Db, schemaName, tableName, pkColName, input, id,
		[]string{"last_user_update_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartialWithExtras(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id,
		[]string{"last_user_update_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}

func (s Store) VerifyEmail(ctx context.Context, id int64) error {

	stmt := fmt.Sprintf(`UPDATE %s.%s SET email_verified = true WHERE id = $1`, schemaName, tableName)
	_, err := s.Db.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("s.Db.Exec failed: %w", err)
	}

	return nil
}
