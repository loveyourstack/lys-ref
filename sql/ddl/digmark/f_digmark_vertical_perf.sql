DROP FUNCTION IF EXISTS digmark.vertical_perf;
CREATE OR REPLACE FUNCTION digmark.vertical_perf (
	_days_before int
)
RETURNS TABLE (
  start_day date,
  end_day date,

  clicks int,
  conversions int,
  impressions int,
  profit_eur numeric,
  return_on_investment numeric,
  revenue_eur numeric,
  spend_eur numeric,
  vertical text
) AS
$BODY$

SELECT
  current_date -$1 +1 AS start_day,
  current_date AS end_day,

  COALESCE(SUM(dm_cp.clicks),0) AS clicks,
  COALESCE(SUM(dm_cp.conversions),0) AS conversions,
  COALESCE(SUM(dm_cp.impressions),0) AS impressions,
  COALESCE(SUM(dm_cp.profit_eur),0) AS profit_eur,
  CASE WHEN COALESCE(SUM(dm_cp.spend_eur),0) = 0 THEN 0.0 
    ELSE COALESCE(SUM(dm_cp.revenue_eur),0)/COALESCE(SUM(dm_cp.spend_eur),0) - 1 END AS return_on_investment,
  COALESCE(SUM(dm_cp.revenue_eur),0) AS revenue_eur,
  COALESCE(SUM(dm_cp.spend_eur),0) AS spend_eur,
  dm_v.name AS vertical
FROM digmark.campaign_performance dm_cp
JOIN digmark.campaign dm_c ON dm_cp.campaign_fk = dm_c.id
JOIN digmark.vertical dm_v ON dm_c.vertical_fk = dm_v.id
WHERE dm_cp.day_cet > current_date -$1
GROUP BY dm_v.name;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
