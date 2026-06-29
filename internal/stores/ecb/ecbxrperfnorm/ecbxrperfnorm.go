package ecbxrperfnorm

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/perfperiod"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurr"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "ECB normalized exchange rate performance"
	schemaName     string = "ecb"
	tableName      string = "xr_perf_normalized"
	viewName       string = "xr_perf_normalized"
	pkColName      string = "id"
	defaultOrderBy string = "period, from_currency_code, to_currency_code"
)

type Model struct {
	Id               int64            `db:"id" json:"id,omitempty"`
	CreatedAt        lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	Day              lystype.Date     `db:"day" json:"day,omitzero"`
	FromCurrencyFk   int64            `db:"from_currency_fk" json:"from_currency_fk,omitempty"`
	FromCurrencyCode string           `db:"from_currency_code" json:"from_currency_code,omitempty"`
	NormalizedPerf   float64          `db:"normalized_perf" json:"normalized_perf,omitempty"`
	Period           perfperiod.Enum  `db:"period" json:"period,omitempty"`
	Rate             float64          `db:"rate" json:"rate,omitempty"`
	ToCurrencyFk     int64            `db:"to_currency_fk" json:"to_currency_fk,omitempty"`
	ToCurrencyCode   string           `db:"to_currency_code" json:"to_currency_code,omitempty"`
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

func (s Store) Create(ctx context.Context, logger *slog.Logger) error {

	currstore := ecbcurr.Store{Db: s.Db}
	eurCurr, err := currstore.SelectByCode(ctx, "EUR")
	if err != nil {
		return fmt.Errorf("currstore.SelectByCode failed for EUR: %w", err)
	}

	for _, p := range perfperiod.Xr {

		daysBefore, daysAfter, err := perfperiod.Days(p, time.Now())
		if err != nil {
			return fmt.Errorf("perfperiod.Days failed: %w", err)
		}

		rowsDeleted, rowsInserted, err := s.createByPeriod(ctx, p, eurCurr.Id, eurCurr.Code, daysBefore, daysAfter)
		if err != nil {
			return fmt.Errorf("s.createByPeriod failed for period %v: %w", p, err)
		}
		logger.Debug("created xr norm perf records", slog.String("period", p.String()), slog.Int64("rowsDeleted", rowsDeleted), slog.Int64("rowsInserted", rowsInserted))
	}

	// check that all periods contain an equal number of days for each currency

	type uneven struct {
		Period         perfperiod.Enum `db:"period"`
		ToCurrencyCode string          `db:"to_currency_code"`
		Count          int64           `db:"cnt"`
	}

	stmt := "SELECT * FROM ecb.v_xr_perf_uneven;"
	unevens, err := lyspg.SelectT[uneven](ctx, s.Db, stmt)
	if err != nil {
		return fmt.Errorf("lyspg.SelectRaw failed for uneven counts: %w", err)
	}
	if len(unevens) > 0 {
		for _, u := range unevens {
			logger.Warn(fmt.Sprintf("%v: %s: %v", u.Period, u.ToCurrencyCode, u.Count))
		}
		return fmt.Errorf("uneven day counts for 1+ currencies")
	}

	return nil
}

func (s Store) createByPeriod(ctx context.Context, period perfperiod.Enum, fromCurrId int64, fromCurrCode string, daysBefore, daysAfter int) (rowsDeleted, rowsInserted int64, err error) {

	// begin tx
	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("s.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// delete existing records for the period
	stmt := fmt.Sprintf(`DELETE FROM %s.%s WHERE period = $1;`, schemaName, tableName)

	cmd, err := tx.Exec(ctx, stmt, period)
	if err != nil {
		return 0, 0, lyserr.Db{Err: fmt.Errorf("tx.Exec (delete) failed: %w", err), Stmt: stmt}
	}
	rowsDeleted = cmd.RowsAffected()

	// insert new records
	stmt = fmt.Sprintf(`INSERT INTO %s.%s (
			"period", day, from_currency_fk, from_currency_code, normalized_perf, rate, to_currency_fk, to_currency_code)
		SELECT 
			'%s', perf_day, %d, '%s', normalized_perf, rate, to_currency_fk, to_currency_code
		FROM ecb.normalized_xr_perf($1, $2, $3);`,
		schemaName, tableName, period, fromCurrId, fromCurrCode)

	cmd, err = tx.Exec(ctx, stmt, fromCurrId, daysBefore, daysAfter)
	if err != nil {
		fmt.Println(stmt)
		return 0, 0, lyserr.Db{Err: fmt.Errorf("tx.Exec (insert) failed: %w", err), Stmt: stmt}
	}
	rowsInserted = cmd.RowsAffected()

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("tx.Commit failed: %w", err)
	}

	return rowsDeleted, rowsInserted, nil
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

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}
