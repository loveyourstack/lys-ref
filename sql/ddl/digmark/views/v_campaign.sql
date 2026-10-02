
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
