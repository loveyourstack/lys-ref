
CREATE OR REPLACE VIEW digmark.v_vertical AS
  SELECT
    dm_v.id,
    dm_v.created_at,
    dm_v.name,
    dm_v.updated_at,
    COALESCE(dm_c.cnt, 0) AS campaign_count
  FROM digmark.vertical dm_v
  LEFT JOIN (SELECT vertical_fk, count(*) AS cnt FROM digmark.campaign GROUP BY vertical_fk) dm_c ON dm_v.id = dm_c.vertical_fk;
