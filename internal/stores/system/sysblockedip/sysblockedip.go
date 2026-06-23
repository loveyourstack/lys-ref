package sysblockedip

import (
	"context"
	"log"
	"net/netip"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "System blocked IPs"
	schemaName     string = "system"
	tableName      string = "blocked_ip"
	viewName       string = "blocked_ip"
	pkColName      string = "id"
	defaultOrderBy string = "created_at DESC"
)

type Input struct {
	Ip netip.Addr `db:"ip" json:"ip,omitzero" validate:"required"`
}

type Model struct {
	Id        int64            `db:"id" json:"id,omitempty"`
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

func (s Store) Delete(ctx context.Context, id int64) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) DeleteByIp(ctx context.Context, ip netip.Addr) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, "ip", ip)
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
