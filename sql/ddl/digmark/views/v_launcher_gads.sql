
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
