
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
