package pubbookarch

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubbook"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Archived books"
	schemaName     string = "publisher"
	tableName      string = "book_archived"
	viewName       string = "v_book_archived"
	defaultOrderBy string = "author, name"
)

type Model struct {
	ArchivedAt        lystype.Datetime `db:"archived_at" json:"archived_at,omitzero"`
	ArchivedByCascade bool             `db:"archived_by_cascade" json:"archived_by_cascade"`
	AuthorIsArchived  bool             `db:"author_is_archived" json:"author_is_archived"`
	pubbook.Model
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

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}
