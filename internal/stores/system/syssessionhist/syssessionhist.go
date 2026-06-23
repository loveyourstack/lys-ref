package syssessionhist

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Session history"
	schemaName     string = "system"
	tableName      string = "session_history"
	viewName       string = "session_history"
	defaultOrderBy string = "last_access_at DESC"
)

type Model struct {
	Id         int64            `db:"id" json:"id,omitempty"`
	ArchivedAt lystype.Datetime `db:"archived_at" json:"archived_at,omitzero"`
	syssession.Model
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) BulkInsert(ctx context.Context, inputs []syssession.Input) (rowsAffected int64, err error) {
	return lyspg.BulkInsert(ctx, s.Db, schemaName, tableName, inputs)
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
