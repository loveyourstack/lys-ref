package procpoint

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/runstatus"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Process point"
	schemaName     string = "process"
	tableName      string = "point"
	viewName       string = "point"
	pkColName      string = "id"
	defaultOrderBy string = "run_fk, step_id"
)

type Input struct {
	Cmd          string           `db:"cmd" json:"cmd,omitempty" validate:"required,max=255"`
	DependsOn    []int64          `db:"depends_on" json:"depends_on,omitempty"`
	DisplayOrder int              `db:"display_order" json:"display_order,omitempty" validate:"required,min=1"`
	ErrMsg       string           `db:"err_msg" json:"err_msg,omitempty"`
	FinishedAt   lystype.Datetime `db:"finished_at" json:"finished_at,omitzero"`
	RunFk        int64            `db:"run_fk" json:"run_fk,omitempty" validate:"required"`
	StartedAt    lystype.Datetime `db:"started_at" json:"started_at,omitzero"`
	Status       runstatus.Enum   `db:"status" json:"status,omitempty" validate:"required,max=255"`
	StepId       int64            `db:"step_id" json:"step_id,omitempty" validate:"required"`
	StepName     string           `db:"step_name" json:"step_name,omitempty" validate:"required"`
}

type Model struct {
	Id        int64            `db:"id" json:"id"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
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

func BulkInsertTx(ctx context.Context, tx pgx.Tx, inputs []Input) (rowsAffected int64, err error) {
	return lyspg.BulkInsert(ctx, tx, schemaName, tableName, inputs)
}

func (s Store) CancelByRunId(ctx context.Context, runId int64) (err error) {

	// set Waiting points to Cancelled
	stmt := fmt.Sprintf("UPDATE %s.%s SET status = 'Cancelled' WHERE run_fk = $1 AND status = 'Waiting';", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, runId)
	if err != nil {
		return fmt.Errorf("s.Db.Exec (Waiting) failed: %w", err)
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

func SelectByRunIdTx(ctx context.Context, tx pgx.Tx, runId int64) (items []Model, err error) {
	stmt := fmt.Sprintf("SELECT * FROM %s.%s WHERE run_fk = $1;", schemaName, viewName)
	return lyspg.SelectT[Model](ctx, tx, stmt, runId)
}

// SelectStatusMapByRunId returns map of k = item.Status, v = []Items
func (s Store) SelectStatusMapByRunId(ctx context.Context, runId int64) (statusMap map[runstatus.Enum][]Model, err error) {
	stmt := fmt.Sprintf("SELECT * FROM %s.%s WHERE run_fk = $1;", schemaName, viewName)
	items, err := lyspg.SelectT[Model](ctx, s.Db, stmt, runId)
	if err != nil {
		return nil, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	statusMap = make(map[runstatus.Enum][]Model)
	for _, item := range items {
		statusMap[item.Status] = append(statusMap[item.Status], item)
	}

	return statusMap, nil
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SetError(ctx context.Context, errMsg string, id int64) (err error) {

	stmt := fmt.Sprintf("UPDATE %s.%s SET finished_at = now(), status = 'Error', err_msg = $1 WHERE id = $2;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, errMsg, id)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("s.Db.Exec: %w", err), Stmt: stmt}
	}

	return nil
}

func (s Store) SetCompleted(ctx context.Context, id int64) (err error) {

	stmt := fmt.Sprintf("UPDATE %s.%s SET finished_at = now(), status = 'Completed' WHERE id = $1;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, id)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("s.Db.Exec: %w", err), Stmt: stmt}
	}

	return nil
}

func (s Store) SetInterrupted(ctx context.Context, id int64) (err error) {

	stmt := fmt.Sprintf("UPDATE %s.%s SET finished_at = now(), status = 'Interrupted' WHERE id = $1;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, id)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("s.Db.Exec: %w", err), Stmt: stmt}
	}

	return nil
}

func (s Store) SetRunning(ctx context.Context, id int64) (err error) {

	stmt := fmt.Sprintf("UPDATE %s.%s SET started_at = now(), status = 'Running' WHERE id = $1;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, id)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("s.Db.Exec: %w", err), Stmt: stmt}
	}

	return nil
}

func (s Store) SetStatus(ctx context.Context, id int64, newStatus runstatus.Enum) (err error) {

	stmt := fmt.Sprintf("UPDATE %s.%s SET status = $1 WHERE id = $2;", schemaName, tableName)
	_, err = s.Db.Exec(ctx, stmt, newStatus, id)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("s.Db.Exec: %w", err), Stmt: stmt}
	}

	return nil
}

func (s Store) Update(ctx context.Context, input Input, id int64) (err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id)
}

func UpdateDependsOnIdsTx(ctx context.Context, tx pgx.Tx, runId int64) (numUpdated int, err error) {

	// select all points for this run
	items, err := SelectByRunIdTx(ctx, tx, runId)
	if err != nil {
		return 0, fmt.Errorf("SelectByRunIdTx failed: %w", err)
	}

	// make map of k = step id, v = point id
	idMap := make(map[int64]int64)
	for _, item := range items {
		idMap[item.StepId] = item.Id
	}

	// for each point
	count := 0
	for _, item := range items {

		// skip if no dependant ids
		if len(item.DependsOn) == 0 {
			continue
		}

		// replace step fks with point ids
		newDependsOn := make([]int64, len(item.DependsOn))
		for i, stepFk := range item.DependsOn {
			newFk, ok := idMap[stepFk]
			if !ok {
				return 0, fmt.Errorf("new fk not found for stepFk %v", stepFk)
			}
			newDependsOn[i] = newFk
		}

		stmt := fmt.Sprintf("UPDATE %s.%s SET depends_on = $1 WHERE id = $2;", schemaName, tableName)
		_, err = tx.Exec(ctx, stmt, newDependsOn, item.Id)
		if err != nil {
			return 0, lyserr.Db{Err: fmt.Errorf("tx.Exec failed on id: %v: %w", item.Id, err), Stmt: stmt}
		}
		count++
	}

	return count, nil
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
