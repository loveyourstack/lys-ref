package dmcampperfagg

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/perfperiod"
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
	Id                 int64            `db:"id" json:"id,omitzero"`
	CampaignFk         int64            `db:"campaign_fk" json:"campaign_fk,omitzero"`
	Clicks             int              `db:"clicks" json:"clicks"`
	Conversions        int              `db:"conversions" json:"conversions"`
	CreatedAt          lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	EndDay             lystype.Date     `db:"end_day" json:"end_day,omitzero"`
	Impressions        int              `db:"impressions" json:"impressions"`
	Period             perfperiod.Enum  `db:"period" json:"period,omitzero"`
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

func (s Store) Create(ctx context.Context, logger *slog.Logger) error {

	/*
	  shows how to capture and handle PostgreSQL notices using pgx for demonstration purposes
	  would normally be used in a more complex SQL process than this
	*/

	var notices []*pgconn.Notice

	// copy the store connection config and set a notice handler
	cc := s.Db.Config().ConnConfig.Copy()
	cc.OnNotice = func(conn *pgconn.PgConn, notice *pgconn.Notice) {
		notices = append(notices, notice)
	}

	// make a connection using the modified config
	conn, err := pgx.ConnectConfig(ctx, cc)
	if err != nil {
		return fmt.Errorf("pgx.ConnectConfig failed: %w", err)
	}
	defer conn.Close(ctx)

	// begin tx to ensure periods are aggregated atomically
	// this is ok because p_aggregate_campaign_perf_by_period has no tx control of its own
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("conn.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// capture Now() outside loop to ensure consistent reference time for all periods
	baseTime := time.Now()

	// prepare tx statement outside loop
	_, err = tx.Prepare(ctx, "agg_camp_perf", fmt.Sprintf("CALL %s.p_aggregate_campaign_perf_by_period($1, $2, $3)", schemaName))
	if err != nil {
		return fmt.Errorf("tx.Prepare failed: %w", err)
	}

	for _, p := range perfperiod.All {

		daysBefore, daysAfter, err := perfperiod.Days(p, baseTime)
		if err != nil {
			return fmt.Errorf("perfperiod.Days failed on period: %s: %w", p, err)
		}

		// call prepared statement for this period using tx from notice-handling connection
		_, err = tx.Exec(ctx, "agg_camp_perf", p, daysBefore, daysAfter)
		if err != nil {
			return fmt.Errorf("tx.Exec failed on period: %s: %w", p, err)
		}
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit failed: %w", err)
	}

	// print notices
	for _, notice := range notices {
		logger.Debug("camp perf agg",
			slog.String("severity", notice.Severity),
			slog.String("message", notice.Message))
	}

	return nil
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
