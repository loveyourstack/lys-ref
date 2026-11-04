
-- for smoke test
INSERT INTO digmark.campaign_performance_aggregated (
  campaign_fk,
  clicks,
  conversions,
  end_day,
  impressions,
  period,
  revenue_eur,
  spend_eur,
  start_day,
  trend,
  volatility
) VALUES (1, 10, 2, '2024-06-01', 100, 'Today', 100.00, 50.00, '2024-06-01', 0.5, 0.1);
