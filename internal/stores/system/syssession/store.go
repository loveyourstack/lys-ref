package syssession

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
	name           string = "System sessions"
	schemaName     string = "system"
	tableName      string = "session"
	viewName       string = "session"
	pkColName      string = "id"
	defaultOrderBy string = "last_access_at DESC"
)

type Input struct {
	AllowMultipleSessions bool             `db:"allow_multiple_sessions" json:"allow_multiple_sessions,omitempty"`
	CreatedAt             lystype.Datetime `db:"created_at" json:"created_at,omitzero" validate:"required"`
	Email                 string           `db:"email" json:"email,omitempty" validate:"required,email,max=256"`
	ExpiresAt             lystype.Datetime `db:"expires_at" json:"expires_at,omitzero" validate:"required"`
	FamilyName            string           `db:"family_name" json:"family_name,omitempty" validate:"required,max=256"`
	ForcePasswordChange   bool             `db:"force_password_change" json:"force_password_change,omitempty"`
	GivenName             string           `db:"given_name" json:"given_name,omitempty" validate:"required,max=256"`
	GeoIpCountryIsoCode   string           `db:"geo_ip_country_iso_code" json:"geo_ip_country_iso_code,omitempty" validate:"required,len=2"`
	GeoIpLocation         string           `db:"geo_ip_location" json:"geo_ip_location,omitempty" validate:"required,max=256"`
	Ip                    netip.Addr       `db:"ip" json:"ip,omitzero" validate:"required"`
	LastAccessAt          lystype.Datetime `db:"last_access_at" json:"last_access_at,omitzero" validate:"required"`
	Roles                 []string         `db:"roles" json:"roles,omitempty" validate:"required"`
	Token                 string           `db:"token" json:"-" validate:"required,max=64"`
	UserAgent             string           `db:"user_agent" json:"user_agent,omitempty" validate:"required,max=256"`
	UserFk                int64            `db:"user_fk" json:"user_fk,omitempty" validate:"required"`
	UserName              string           `db:"user_name" json:"user_name,omitempty" validate:"required,max=64"`
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
