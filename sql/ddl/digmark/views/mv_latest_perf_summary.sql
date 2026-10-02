
-- using MV for this for demo purposes, would normally not be needed for such a simple query
CREATE MATERIALIZED VIEW digmark.mv_latest_perf_summary AS
  SELECT day_cet AS day, SUM(spend_eur) AS total_spend, SUM(revenue_eur) AS total_revenue 
  FROM digmark.campaign_performance
  WHERE day_cet > current_date -7
  GROUP BY 1 ORDER BY 1;
