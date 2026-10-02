DROP PROCEDURE IF EXISTS digmark.p_aggregate_campaign_perf_by_period;
CREATE OR REPLACE PROCEDURE digmark.p_aggregate_campaign_perf_by_period (
  _period core.performance_period,
  _days_before int,
  _days_after int
) AS 
$BODY$
DECLARE
  _rows_deleted int;
  _rows_inserted int;
BEGIN

/*
  procedure: for ETL and batch operations
  - assuming PL/pgSQL, allows transaction control (not shown below. Could be used if loop over periods were done here)
  - no return value, but can use OUT params (for small number of outputs) or raise notices which can be read by pgx (shown below for demo purposes)
  - executed via CALL, not inside query
  - will most likely not use standard-SQL form (BEGIN ATOMIC), so dependencies will not be tracked
*/

DELETE FROM digmark.campaign_performance_aggregated WHERE period = _period;
GET DIAGNOSTICS _rows_deleted = ROW_COUNT;
RAISE NOTICE 'period %: deleted % existing rows', _period, _rows_deleted;

INSERT INTO digmark.campaign_performance_aggregated (campaign_fk, "period", start_day, end_day, clicks, conversions, impressions, revenue_eur, spend_eur, trend, volatility)
  SELECT campaign_fk, _period, start_day, end_day, clicks, conversions, impressions, revenue_eur, spend_eur, trend, volatility
  FROM digmark.f_aggregate_campaign_perf(_days_before, _days_after);
GET DIAGNOSTICS _rows_inserted = ROW_COUNT;
RAISE NOTICE 'period %: inserted % rows', _period, _rows_inserted;

--  ... transaction control, further operations, notices

END;
$BODY$
LANGUAGE plpgsql;