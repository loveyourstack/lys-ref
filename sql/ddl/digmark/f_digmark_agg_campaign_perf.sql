DROP FUNCTION IF EXISTS digmark.aggregate_campaign_perf;
CREATE OR REPLACE FUNCTION digmark.aggregate_campaign_perf (
	_days_before int,
	_days_after int
)
RETURNS TABLE (
  start_day date,
  end_day date,
  
  campaign_fk bigint,
  clicks int,
  conversions int,
  impressions int,
  revenue_eur numeric,
  spend_eur numeric,
  trend numeric,
  volatility numeric
) AS
$BODY$

WITH day_seq AS (
  -- for each campaign, first day is 1, second day is 2, etc. Used for slope func below
  SELECT campaign_fk, day_cet, row_number() OVER (PARTITION BY campaign_fk ORDER BY day_cet) AS num
  FROM digmark.campaign_performance
  WHERE day_cet BETWEEN current_date +$1 AND current_date +$2
)
SELECT
  current_date +$1 AS start_day,
  current_date +$2 AS end_day,
  
  dm_cp.campaign_fk,
  SUM(dm_cp.clicks) AS clicks,
  SUM(dm_cp.conversions) AS conversions,
  SUM(dm_cp.impressions) AS impressions,
  SUM(dm_cp.revenue_eur) AS revenue_eur,
  SUM(dm_cp.spend_eur) AS spend_eur,
  CASE WHEN $2 > $1 AND COUNT(*)::numeric / ((current_date +$2) - (current_date +$1)) > 0.66 THEN 
    COALESCE(regr_slope((dm_cp.revenue_eur - dm_cp.spend_eur), day_seq.num)::numeric,0.0) ELSE 0 
  END as trend,
  CASE WHEN $2 > $1 AND COUNT(*)::numeric / ((current_date +$2) - (current_date +$1)) > 0.66 AND SUM(dm_cp.spend_eur) > 0 THEN 
    COALESCE(stddev(dm_cp.revenue_eur - dm_cp.spend_eur) / SUM(dm_cp.spend_eur),0.0) ELSE 0 
  END AS volatility
FROM digmark.campaign_performance dm_cp
JOIN day_seq USING (campaign_fk, day_cet)
WHERE dm_cp.day_cet BETWEEN current_date +$1 AND current_date +$2
GROUP BY dm_cp.campaign_fk;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
