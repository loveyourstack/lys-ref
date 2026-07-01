
CREATE OR REPLACE VIEW digmark.v_campaign AS
  SELECT
    dm_c.id,
    dm_c.country_fk,
      geo_c.iso2 AS country_iso2,
      geo_c.name AS country,
    dm_c.created_at,
    dm_c.daily_budget_eur,
    dm_c.is_active,
    dm_c.manager,
    dm_c.name,
    dm_c.updated_at,
    dm_c.vertical_fk,
      dm_v.name AS vertical,
    CASE WHEN dm_cp.min_day IS NULL OR dm_cp.max_day IS NULL THEN 'No performance' 
      ELSE to_char(dm_cp.min_day, 'DD Mon YYYY') || ' - ' || to_char(dm_cp.max_day, 'DD Mon YYYY') END AS performance_range
  FROM digmark.campaign dm_c
  JOIN geo.country geo_c ON dm_c.country_fk = geo_c.id
  JOIN digmark.vertical dm_v ON dm_c.vertical_fk = dm_v.id
  LEFT JOIN (SELECT campaign_fk, MIN(day_cet) AS min_day, MAX(day_cet) AS max_day FROM digmark.campaign_performance GROUP BY campaign_fk) dm_cp ON dm_c.id = dm_cp.campaign_fk;


CREATE OR REPLACE VIEW digmark.v_campaign_performance AS
  SELECT
    dm_cp.id,
    dm_cp.campaign_fk,
      dm_c.name AS campaign,
      dm_c.country_fk,
        geo_c.iso2 AS country_iso2,
        geo_c.name AS country,
      dm_c.vertical_fk,
        dm_v.name AS vertical,
    dm_cp.clicks,
    dm_cp.conversions,
    dm_cp.day_cet,
    dm_cp.impressions,
    dm_cp.revenue_eur,
    dm_cp.spend_eur,
    dm_cp.profit_eur,
    dm_cp.return_on_investment,
    dm_cp.created_at,
    dm_cp.updated_at
  FROM digmark.campaign_performance dm_cp
  JOIN digmark.campaign dm_c ON dm_cp.campaign_fk = dm_c.id
  JOIN geo.country geo_c ON dm_c.country_fk = geo_c.id
  JOIN digmark.vertical dm_v ON dm_c.vertical_fk = dm_v.id;


CREATE OR REPLACE VIEW digmark.v_latest_perf_summary AS
  SELECT day_cet AS day, SUM(spend_eur) AS total_spend, SUM(revenue_eur) AS total_revenue 
  FROM digmark.campaign_performance
  WHERE day_cet > current_date -7
  GROUP BY 1 ORDER BY 1;


CREATE OR REPLACE VIEW digmark.v_launcher_fb AS
  SELECT
    dm_l_fb.id,
    dm_l_fb.country_fk,
      geo_c.iso2 AS country_iso2,
      geo_c.name AS country,
    dm_l_fb.created_at,
    dm_l_fb.created_at_day,
    dm_l_fb.daily_budget_eur,
    dm_l_fb.fan_page,
    dm_l_fb.fb_account_id,
    dm_l_fb.fb_campaign_id,
    dm_l_fb.fb_creative_id,
    dm_l_fb.manager,
    dm_l_fb.max_steps,
    dm_l_fb.message,
    dm_l_fb.name,
    dm_l_fb.partner,
    dm_l_fb.status,
    dm_l_fb.step,
    dm_l_fb.updated_at,
    dm_l_fb.vertical_fk,
      dm_v.name AS vertical
  FROM digmark.launcher_fb dm_l_fb
  JOIN geo.country geo_c ON dm_l_fb.country_fk = geo_c.id
  JOIN digmark.vertical dm_v ON dm_l_fb.vertical_fk = dm_v.id;


CREATE OR REPLACE VIEW digmark.v_launcher_gads AS
  SELECT
    dm_l_gads.id,
    dm_l_gads.country_fk,
      geo_c.iso2 AS country_iso2,
      geo_c.name AS country,
    dm_l_gads.created_at,
    dm_l_gads.created_at_day,
    dm_l_gads.daily_budget_eur,
    dm_l_gads.gads_account_id,
    dm_l_gads.gads_ad_id,
    dm_l_gads.gads_ad_group_id,
    dm_l_gads.gads_campaign_id,
    dm_l_gads.manager,
    dm_l_gads.max_steps,
    dm_l_gads.message,
    dm_l_gads.name,
    dm_l_gads.partner,
    dm_l_gads.status,
    dm_l_gads.step,
    dm_l_gads.updated_at,
    dm_l_gads.vertical_fk,
      dm_v.name AS vertical
  FROM digmark.launcher_gads dm_l_gads
  JOIN geo.country geo_c ON dm_l_gads.country_fk = geo_c.id
  JOIN digmark.vertical dm_v ON dm_l_gads.vertical_fk = dm_v.id;


CREATE OR REPLACE VIEW digmark.v_vertical AS
  SELECT
    dm_v.id,
    dm_v.created_at,
    dm_v.name,
    dm_v.updated_at,
    COALESCE(dm_c.cnt, 0) AS campaign_count
  FROM digmark.vertical dm_v
  LEFT JOIN (SELECT vertical_fk, count(*) AS cnt FROM digmark.campaign GROUP BY vertical_fk) dm_c ON dm_v.id = dm_c.vertical_fk;
  