DROP FUNCTION IF EXISTS digmark.daily_trend;
CREATE OR REPLACE FUNCTION digmark.daily_trend (
	_days_before int
)
RETURNS TABLE (
  day date,
  
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
  dm_cp.day_cet AS day,
  
  COALESCE(SUM(dm_cp.clicks),0) AS clicks,
  COALESCE(SUM(dm_cp.conversions),0) AS conversions,
  COALESCE(SUM(dm_cp.impressions),0) AS impressions,
  COALESCE(SUM(dm_cp.profit_eur),0) AS profit_eur,
  CASE WHEN COALESCE(SUM(dm_cp.spend_eur),0) = 0 THEN 0.0 
    ELSE COALESCE(SUM(dm_cp.revenue_eur),0)/COALESCE(SUM(dm_cp.spend_eur),0) - 1 END AS return_on_investment,
  COALESCE(SUM(dm_cp.revenue_eur),0) AS revenue_eur,
  COALESCE(SUM(dm_cp.spend_eur),0) AS spend_eur
FROM digmark.campaign_performance dm_cp
WHERE dm_cp.day_cet > current_date -$1
GROUP BY dm_cp.day_cet;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
