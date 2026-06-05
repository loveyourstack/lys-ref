package pubbook

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Books"
	schemaName     string = "publisher"
	tableName      string = "book"
	viewName       string = "v_book"
	pkColName      string = "id"
	defaultOrderBy string = "author, name"
)

type Input struct {
	AuthorFk int64  `db:"author_fk" json:"author_fk,omitempty" validate:"required"`
	Name     string `db:"name" json:"name,omitempty" validate:"required,max=256"`
}

type Model struct {
	Id               int64            `db:"id" json:"id,omitempty"`
	Author           string           `db:"author" json:"author,omitempty"`
	CreatedAt        lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CreatedBy        string           `db:"created_by" json:"created_by,omitempty"`                   // assigned in Insert func
	LastUserUpdateBy string           `db:"last_user_update_by" json:"last_user_update_by,omitempty"` // assigned in Update funcs
	UpdatedAt        lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"`                    // assigned by trigger
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

func (s Store) Archive(ctx context.Context, tx pgx.Tx, id int64) error {
	return lyspg.Archive(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) ArchiveCascadedByAuthor(ctx context.Context, tx pgx.Tx, authorId int64) error {
	return lyspg.Archive(ctx, tx, schemaName, tableName, "author_fk", authorId, true)
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
	return lyspg.InsertWithExtras[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input, []string{"created_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) Restore(ctx context.Context, tx pgx.Tx, id int64) error {
	return lyspg.Restore(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) RestoreCascadedByAuthor(ctx context.Context, tx pgx.Tx, authorId int64) error {
	return lyspg.Restore(ctx, tx, schemaName, tableName, "author_fk", authorId, true)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.UpdateWithExtras(ctx, s.Db, schemaName, tableName, pkColName, input, id, []string{"last_user_update_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartialWithExtras(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id,
		[]string{"last_user_update_by"}, []any{lys.GetUserNameFromCtx(ctx)})
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
