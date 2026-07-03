package dmlaunch

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/launchstatus"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Digmark campaign launchers"
	schemaName     string = "digmark"
	tableName      string = "launcher"
	viewName       string = "launcher"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

// abstract table containing no records. Only selection allowed. Selection returns child table records

// Input contains the shared input fields for all launchers.
type Input struct {
	DailyBudgetEur float64 `db:"daily_budget_eur" json:"daily_budget_eur,omitempty" validate:"gte=0,lte=10000"`
	Manager        string  `db:"manager" json:"manager,omitempty" validate:"required,max=64"`
	Name           string  `db:"name" json:"name,omitempty" validate:"required,max=256"`
}

// Computed contains the shared computed fields for all launchers.
type Computed struct {
	CountryFk  int64             `db:"country_fk" json:"country_fk,omitempty"`   // set during preparation
	Message    string            `db:"message" json:"message,omitempty"`         // set during preparation and processing
	Status     launchstatus.Enum `db:"status" json:"status,omitempty"`           // set during preparation and processing
	Step       int               `db:"step" json:"step,omitempty"`               // set during processing
	VerticalFk int64             `db:"vertical_fk" json:"vertical_fk,omitempty"` // set during preparation
}

// DbManaged contains the shared database-managed fields for all launchers.
type DbManaged struct {
	Id           int64            `db:"id" json:"id,omitempty"`
	Country      string           `db:"country" json:"country,omitempty"`
	CountryIso2  string           `db:"country_iso2" json:"country_iso2,omitempty"`
	CreatedAt    lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	CreatedAtDay lystype.Date     `db:"created_at_day" json:"created_at_day,omitzero"`
	MaxSteps     int              `db:"max_steps" json:"max_steps,omitempty"`
	Partner      string           `db:"partner" json:"partner,omitempty"`
	UpdatedAt    lystype.Datetime `db:"updated_at" json:"updated_at,omitzero"` // assigned by trigger
	Vertical     string           `db:"vertical" json:"vertical,omitempty"`
}

type Model struct {
	Input
	Computed
	DbManaged
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

type statusCheck struct {
	Status launchstatus.Enum `db:"status"`
}
type idStatusCheck struct {
	Id     int64             `db:"id"`
	Status launchstatus.Enum `db:"status"`
}

func CancelMany(ctx context.Context, db *pgxpool.Pool, pSchema, pTable string, ids []int64) (numDeleted int64, err error) {

	// select items from ids
	items, err := lyspg.SelectT[idStatusCheck](ctx, db, fmt.Sprintf("SELECT id, status FROM %s.%s WHERE id = ANY($1);", pSchema, pTable), ids)
	if err != nil {
		return 0, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	var cancelableIds []int64

	// only allow cancellation of records where status is Queued, ignore others
	for _, item := range items {
		if item.Status == launchstatus.Queued {
			cancelableIds = append(cancelableIds, item.Id)
		}
	}

	// set back to Ready
	stmt := fmt.Sprintf("UPDATE %s.%s SET status = 'Ready' WHERE id = ANY($1);", pSchema, pTable)
	res, err := db.Exec(ctx, stmt, cancelableIds)
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("db.Exec failed: %w", err), Stmt: stmt}
	}

	return res.RowsAffected(), nil
}

// ClaimForPreparation selects a batch of unprepared items and sets their status to Preparing, returning the items.
func ClaimForPreparation[T any](ctx context.Context, db *pgxpool.Pool, pSchema, pTable string, batchSize int) (items []T, err error) {
	stmt := fmt.Sprintf(`WITH picked AS (SELECT id FROM %s.%s WHERE status = 'Unprepared' ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED) 
		UPDATE %s.%s l SET status = 'Preparing', message = '' FROM picked 
		WHERE l.id = picked.id RETURNING l.*;`, pSchema, pTable, pSchema, pTable)
	return lyspg.SelectT[T](ctx, db, stmt, batchSize)
}

// ClaimNextForProcessing selects the next queued item and sets its status to Processing, returning the item.
func ClaimNextForProcessing[T any](ctx context.Context, db *pgxpool.Pool, pSchema, pTable string) (item T, err error) {
	stmt := fmt.Sprintf(`WITH picked AS (SELECT id FROM %s.%s WHERE status = 'Queued' ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED) 
		UPDATE %s.%s l SET status = 'Processing', message = '' FROM picked 
		WHERE l.id = picked.id RETURNING l.*;`, pSchema, pTable, pSchema, pTable)
	items, err := lyspg.SelectT[T](ctx, db, stmt)
	var empty T
	if err != nil {
		return empty, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}
	if len(items) == 0 {
		return empty, pgx.ErrNoRows
	}
	return items[0], nil
}

// DeleteEditableByID deletes a row only when status is editable.
func DeleteEditableByID(ctx context.Context, db *pgxpool.Pool, pSchema, pTable, pView, pPkCol string, id int64) error {

	// select status from id
	item, err := lyspg.SelectUniqueRowFields[statusCheck](ctx, db, []string{"status"}, pSchema, pView, pPkCol, id)
	if err != nil {
		return fmt.Errorf("lyspg.SelectUniqueRowField failed: %w", err)
	}

	// reject if status does not allow deletion
	if !slices.Contains(launchstatus.Editable[:], item.Status) {
		return lyserr.User{Message: fmt.Sprintf("deletion not allowed for status: %s", item.Status)}
	}

	return lyspg.DeleteUnique(ctx, db, pSchema, pTable, pPkCol, id)
}

func DeleteMany(ctx context.Context, db *pgxpool.Pool, pSchema, pTable string, ids []int64) (numDeleted int64, err error) {

	// select items from ids
	items, err := lyspg.SelectT[idStatusCheck](ctx, db, fmt.Sprintf("SELECT id, status FROM %s.%s WHERE id = ANY($1);", pSchema, pTable), ids)
	if err != nil {
		return 0, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	var deleteableIds []int64

	// only allow deletion of records where status allows it, ignore others
	for _, item := range items {
		if slices.Contains(launchstatus.Editable[:], item.Status) {
			deleteableIds = append(deleteableIds, item.Id)
		}
	}

	// delete
	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE id = ANY($1);", pSchema, pTable)
	res, err := db.Exec(ctx, stmt, deleteableIds)
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("db.Exec failed: %w", err), Stmt: stmt}
	}

	return res.RowsAffected(), nil
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func QueueMany(ctx context.Context, db *pgxpool.Pool, pSchema, pTable string, ids []int64) (numDeleted int64, err error) {

	// select items from ids
	items, err := lyspg.SelectT[idStatusCheck](ctx, db, fmt.Sprintf("SELECT id, status FROM %s.%s WHERE id = ANY($1);", pSchema, pTable), ids)
	if err != nil {
		return 0, fmt.Errorf("lyspg.SelectT failed: %w", err)
	}

	// only allow queueing of Ready records, ignore others
	var queueableIds []int64
	for _, item := range items {
		if item.Status == launchstatus.Ready {
			queueableIds = append(queueableIds, item.Id)
		}
	}

	// queue
	stmt := fmt.Sprintf("UPDATE %s.%s SET status = 'Queued' WHERE id = ANY($1);", pSchema, pTable)
	res, err := db.Exec(ctx, stmt, queueableIds)
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("db.Exec failed: %w", err), Stmt: stmt}
	}

	return res.RowsAffected(), nil
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}
