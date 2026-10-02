
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
