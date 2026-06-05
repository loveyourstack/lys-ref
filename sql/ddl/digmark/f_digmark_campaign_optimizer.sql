DROP FUNCTION IF EXISTS digmark.campaign_optimizer;
CREATE OR REPLACE FUNCTION digmark.campaign_optimizer (
	_period core.performance_period
)
RETURNS TABLE (
  start_day date,
  end_day date,
  
  id bigint,
  country_fk bigint,
    country text,
    country_iso2 text,
  daily_budget_eur numeric,
  is_active boolean,
  manager digmark.manager,
  name text,
  vertical_fk bigint,
    vertical text,
  
  clicks int,
  conversions int,
  impressions int,
  profit_eur numeric,
  revenue_eur numeric,
  return_on_investment numeric,
  spend_eur numeric,
  trend numeric,
  volatility numeric
) AS
$BODY$

WITH agg_perf AS (
  SELECT *
  FROM digmark.campaign_performance_aggregated
  WHERE period = _period
)
SELECT
  COALESCE(agg_perf.start_day,'0001-01-01') AS start_day,
  COALESCE(agg_perf.end_day,'0001-01-01') AS end_day,
  
  dm_c.id,
  dm_c.country_fk,
    geo_c.name AS country,
    geo_c.iso2 AS country_iso2,
  dm_c.daily_budget_eur,
  dm_c.is_active,
  dm_c.manager,
  dm_c.name,
  dm_c.vertical_fk,
    dm_v.name AS vertical,
  
  COALESCE(agg_perf.clicks,0) AS clicks,
  COALESCE(agg_perf.conversions,0) AS conversions,
  COALESCE(agg_perf.impressions,0) AS impressions,
  COALESCE(agg_perf.profit_eur,0) AS profit_eur,
  COALESCE(agg_perf.revenue_eur,0) AS revenue_eur,
  COALESCE(agg_perf.return_on_investment,0) AS return_on_investment,
  COALESCE(agg_perf.spend_eur,0) AS spend_eur,
  COALESCE(agg_perf.trend,0) AS trend,
  COALESCE(agg_perf.volatility,0) AS volatility

FROM digmark.campaign dm_c
JOIN geo.country geo_c ON dm_c.country_fk = geo_c.id
JOIN digmark.vertical dm_v ON dm_c.vertical_fk = dm_v.id
LEFT JOIN agg_perf ON dm_c.id = agg_perf.campaign_fk;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
