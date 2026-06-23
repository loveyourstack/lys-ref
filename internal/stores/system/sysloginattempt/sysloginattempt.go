package sysloginattempt

import (
	"context"
	"fmt"
	"log"
	"net/netip"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "System login attempts"
	schemaName     string = "system"
	tableName      string = "login_attempt"
	viewName       string = "login_attempt"
	pkColName      string = "id"
	defaultOrderBy string = "created_at DESC"
)

type Input struct {
	CreatedAt   lystype.Datetime `db:"created_at" json:"created_at,omitzero" validate:"required"`
	Ip          netip.Addr       `db:"ip" json:"ip,omitzero" validate:"required"`
	IsBlocked   bool             `db:"is_blocked" json:"is_blocked"`
	NumAttempts int              `db:"num_attempts" json:"num_attempts"`
}

type Model struct {
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

func (s Store) BulkInsertTx(ctx context.Context, tx pgx.Tx, inputs []Input) (rowsAffected int64, err error) {
	return lyspg.BulkInsert(ctx, tx, schemaName, tableName, inputs)
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectAll(ctx context.Context) (items []Model, err error) {
	items, _, err = lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), lyspg.SelectParams{Limit: -1})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) TruncateTx(ctx context.Context, tx pgx.Tx) error {
	stmt := fmt.Sprintf("TRUNCATE TABLE %s.%s;", schemaName, tableName)
	_, err := tx.Exec(ctx, stmt)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("tx.Exec failed: %w", err), Stmt: stmt}
	}

	return nil
}
