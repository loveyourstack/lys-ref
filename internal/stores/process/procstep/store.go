package procstep

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Process step"
	schemaName     string = "process"
	tableName      string = "step"
	viewName       string = "v_step"
	pkColName      string = "id"
	defaultOrderBy string = "flow_fk, display_order"
)

type Input struct {
	Cmd          string `db:"cmd" json:"cmd,omitempty" validate:"required,max=255"`
	DisplayOrder int    `db:"display_order" json:"display_order,omitempty" validate:"required,min=1"`
	FlowFk       int64  `db:"flow_fk" json:"flow_fk,omitempty" validate:"required"`
	Name         string `db:"name" json:"name,omitempty" validate:"required,max=255"`
}

type Model struct {
	Id             int64            `db:"id" json:"id"`
	CreatedAt      lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	DependsOn      []int64          `db:"depends_on" json:"depends_on,omitempty"`
	DependsOnNames []string         `db:"depends_on_names" json:"depends_on_names,omitempty"`
	Flow           string           `db:"flow" json:"flow,omitempty"`
	PointCount     int              `db:"point_count" json:"point_count,omitempty"`
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

// SelectAvailableDependencies returns the steps which may be added as a dependency without causing a circular dependency or a redundancy
func (s Store) SelectAvailableDependencies(ctx context.Context, id int64) (items []Model, err error) {
	stmt := `WITH RECURSIVE deps AS (
				SELECT depends_on_fk FROM process.step_link WHERE step_fk = $1
				UNION 
				SELECT proc_sl.depends_on_fk FROM process.step_link proc_sl JOIN deps ON proc_sl.step_fk = deps.depends_on_fk
			), uses AS (
				SELECT step_fk FROM process.step_link WHERE depends_on_fk = $1
				UNION 
				SELECT proc_sl.step_fk FROM process.step_link proc_sl JOIN uses ON proc_sl.depends_on_fk = uses.step_fk
			)
			SELECT * FROM process.v_step WHERE flow_fk = (SELECT flow_fk FROM process.step WHERE id = $1)
				AND NOT EXISTS (SELECT 1 FROM deps WHERE deps.depends_on_fk = id) 
			    AND NOT EXISTS (SELECT 1 FROM uses WHERE uses.step_fk = id)
				AND id != $1
				ORDER BY name;`
	return lyspg.SelectT[Model](ctx, s.Db, stmt, id)
}

func (s Store) SelectByIds(ctx context.Context, ids []int64) (items []Model, err error) {
	return lyspg.SelectBySlice[int64, Model](ctx, s.Db, schemaName, viewName, pkColName, ids)
}

func (s Store) SelectDepIds(ctx context.Context, id int64) (depIds []int64, err error) {
	stmt := `WITH RECURSIVE deps AS (
				SELECT depends_on_fk FROM process.step_link WHERE step_fk = $1
				UNION 
				SELECT proc_sl.depends_on_fk FROM process.step_link proc_sl JOIN deps ON proc_sl.step_fk = deps.depends_on_fk
			)
			SELECT depends_on_fk AS step_id FROM deps ORDER BY 1;`
	return lyspg.SelectSlice[int64](ctx, s.Db, stmt, id)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SwapDisplayOrder(ctx context.Context, id1, id2 int64) (err error) {

	item1, err := lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id1)
	if err != nil {
		return fmt.Errorf("lyspg.SelectUnique (id1) failed: %w", err)
	}

	item2, err := lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id2)
	if err != nil {
		return fmt.Errorf("lyspg.SelectUnique (id2) failed: %w", err)
	}

	// begin tx
	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("s.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt := fmt.Sprintf("UPDATE %s.%s SET display_order = $1 WHERE id = $2;", schemaName, tableName)

	_, err = tx.Exec(ctx, stmt, item2.DisplayOrder, id1)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("update (id1) failed: %w", err), Stmt: stmt}
	}

	_, err = tx.Exec(ctx, stmt, item1.DisplayOrder, id2)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("update (id2) failed: %w", err), Stmt: stmt}
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit failed: %w", err)
	}

	return nil
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {

	// disallow bash combinations and pipes
	if strings.Contains(input.Cmd, "&") || strings.Contains(input.Cmd, "|") || strings.Contains(input.Cmd, ";") {
		return lyserr.User{Message: fmt.Sprintf("combinations or pipes are not allowed: %s", input.Cmd)}
	}

	return lysmeta.Validate(validate, input)
}
