package syssvcaccount

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "System service accounts"
	schemaName     string = "system"
	tableName      string = "service_account"
	viewName       string = "service_account"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

type Input struct {
	ApiKeyHash   string           `db:"api_key_hash" json:"api_key_hash,omitzero" validate:"required,len=64"`
	ExpiresAt    lystype.Datetime `db:"expires_at" json:"expires_at,omitzero"`
	IsActive     bool             `db:"is_active" json:"is_active,omitzero"`
	LastAccessAt lystype.Datetime `db:"last_access_at" json:"last_access_at,omitzero"`
	Name         string           `db:"name" json:"name,omitzero" validate:"required,max=64"`
	Roles        []sysrole.Enum   `db:"roles" json:"roles,omitzero" validate:"required"`
}

type Model struct {
	Id        int64            `db:"id" json:"id,omitzero"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
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

func (s Store) Deactivate(ctx context.Context, id int64) error {

	assignmentsMap := make(map[string]any)
	assignmentsMap["is_active"] = false

	err := s.UpdatePartial(ctx, assignmentsMap, id)
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
	return lyspg.Insert[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectByApiKeyHash(ctx context.Context, apiKeyHash string) (item Model, err error) {

	item, err = lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, "api_key_hash", apiKeyHash)
	if err != nil {
		return item, fmt.Errorf("lyspg.SelectUnique failed: %w", err)
	}

	if !item.IsActive {
		return item, lyserr.User{Message: "Invalid API key", StatusCode: http.StatusUnauthorized}
	}
	if time.Time(item.ExpiresAt).Before(time.Now()) {
		return item, lyserr.User{Message: "Invalid API key", StatusCode: http.StatusUnauthorized}
	}

	// update last_access_at
	assignmentsMap := make(map[string]any)
	assignmentsMap["last_access_at"] = time.Now()
	err = s.UpdatePartial(ctx, assignmentsMap, item.Id)
	if err != nil {
		return item, fmt.Errorf("s.UpdatePartial failed: %w", err)
	}

	return item, nil
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
