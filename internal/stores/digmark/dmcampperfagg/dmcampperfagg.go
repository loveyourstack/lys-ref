package dmcampperfagg

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/perfperiod"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Digmark campaign performance aggregated"
	schemaName     string = "digmark"
	tableName      string = "campaign_performance_aggregated"
	viewName       string = "campaign_performance_aggregated"
	pkColName      string = "id"
	defaultOrderBy string = "period, campaign_fk"
)

type Model struct {
	Id                 int64            `db:"id" json:"id,omitempty"`
	CampaignFk         int64            `db:"campaign_fk" json:"campaign_fk,omitempty"`
	Clicks             int              `db:"clicks" json:"clicks"`
	Conversions        int              `db:"conversions" json:"conversions"`
	CreatedAt          lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	EndDay             lystype.Date     `db:"end_day" json:"end_day,omitzero"`
	Impressions        int              `db:"impressions" json:"impressions"`
	Period             perfperiod.Enum  `db:"period" json:"period,omitempty"`
	ProfitEur          float64          `db:"profit_eur" json:"profit_eur"`
	ReturnOnInvestment float64          `db:"return_on_investment" json:"return_on_investment"`
	RevenueEur         float64          `db:"revenue_eur" json:"revenue_eur"`
	SpendEur           float64          `db:"spend_eur" json:"spend_eur"`
	StartDay           lystype.Date     `db:"start_day" json:"start_day,omitzero"`
	Trend              float64          `db:"trend" json:"trend"`
	Volatility         float64          `db:"volatility" json:"volatility"`
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

func (s Store) Create(ctx context.Context, infoLog *slog.Logger) error {

	for _, p := range perfperiod.All {

		daysBefore, daysAfter, err := perfperiod.Days(p, time.Now())
		if err != nil {
			return fmt.Errorf("perfperiod.Days failed: %w", err)
		}

		rowsDeleted, rowsInserted, err := s.createByPeriod(ctx, p, daysBefore, daysAfter)
		if err != nil {
			return fmt.Errorf("s.createByPeriod failed for period %v: %w", p, err)
		}
		infoLog.Debug("created camp perf agg records", slog.String("period", p.String()), slog.Int64("rowsDeleted", rowsDeleted), slog.Int64("rowsInserted", rowsInserted))
	}

	return nil
}

func (s Store) createByPeriod(ctx context.Context, period perfperiod.Enum, daysBefore, daysAfter int) (rowsDeleted, rowsInserted int64, err error) {

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
			campaign_fk, "period", start_day, end_day,
			clicks, conversions, impressions, revenue_eur, spend_eur, trend, volatility)
		SELECT 
			campaign_fk, '%s', start_day, end_day,
			clicks, conversions, impressions, revenue_eur, spend_eur, trend, volatility
		FROM digmark.aggregate_campaign_perf($1, $2);`,
		schemaName, tableName, period)

	cmd, err = tx.Exec(ctx, stmt, daysBefore, daysAfter)
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
