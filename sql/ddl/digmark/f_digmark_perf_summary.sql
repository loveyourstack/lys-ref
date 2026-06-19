DROP FUNCTION IF EXISTS digmark.perf_summary;
CREATE OR REPLACE FUNCTION digmark.perf_summary (
	_days_before int
)
RETURNS TABLE (
  start_day date,
  end_day date,

  active_campaigns int,
  clicks int,
  conversions int,
  impressions int,
  profit_eur numeric,
  return_on_investment numeric,
  revenue_eur numeric,
  spend_eur numeric
) AS
$BODY$

SELECT
  current_date -$1 +1 AS start_day,
  current_date AS end_day,

  (SELECT COUNT(*) AS cnt FROM digmark.campaign WHERE is_active = true) AS active_campaigns,
  COALESCE(SUM(dm_cp.clicks),0) AS clicks,
  COALESCE(SUM(dm_cp.conversions),0) AS conversions,
  COALESCE(SUM(dm_cp.impressions),0) AS impressions,
  COALESCE(SUM(dm_cp.profit_eur),0) AS profit_eur,
  CASE WHEN COALESCE(SUM(dm_cp.spend_eur),0) = 0 THEN 0.0 
    ELSE COALESCE(SUM(dm_cp.revenue_eur),0)/COALESCE(SUM(dm_cp.spend_eur),0) - 1 END AS return_on_investment,
  COALESCE(SUM(dm_cp.revenue_eur),0) AS revenue_eur,
  COALESCE(SUM(dm_cp.spend_eur),0) AS spend_eur
FROM digmark.campaign_performance dm_cp
WHERE dm_cp.day_cet > current_date -$1;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
