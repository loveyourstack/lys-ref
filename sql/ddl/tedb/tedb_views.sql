
CREATE OR REPLACE VIEW tedb.v_vat_rate AS
  SELECT
    tedb_vr.id,
    tedb_vr.category_fk,
      tedb_vc.identifier AS category_identifier,
    tedb_vr.cn_codes,
    tedb_vr.comment,
    tedb_vr.cpa_codes,
    tedb_vr.created_at,
    tedb_vr.member_state,
      geo_c.name AS country,
    tedb_vr.rate_type,
    tedb_vr.rate,
    tedb_vr.situation_on,
    tedb_vr.type,
    tedb_vr.updated_at
  FROM tedb.vat_rate tedb_vr
  JOIN tedb.vat_category tedb_vc ON tedb_vr.category_fk = tedb_vc.id
  JOIN geo.country geo_c ON tedb_vr.member_state = geo_c.iso2;


CREATE OR REPLACE VIEW tedb.v_vat_rate_summary AS

  WITH latest AS (
    SELECT member_state, type, category_fk, max(situation_on) AS max_situation_on
    FROM tedb.vat_rate
    GROUP BY 1,2,3
  )
  
  SELECT tedb_vr.country, initcap(tedb_vr.type::text) AS type, '' AS categories, situation_on, CASE WHEN comment LIKE '%Canary%' THEN comment ELSE '' END AS comment, rate
  FROM tedb.v_vat_rate tedb_vr
  JOIN latest ON tedb_vr.member_state = latest.member_state AND tedb_vr.type = latest.type AND tedb_vr.category_fk = latest.category_fk
  WHERE tedb_vr.type = 'STANDARD' AND situation_on = latest.max_situation_on
  
  UNION
  
  SELECT tedb_vr.country, initcap(tedb_vr.type::text) AS type, 
    CASE WHEN LENGTH(array_to_string(ARRAY_AGG(DISTINCT(initcap(category_identifier))), ', ')) > 57 
      THEN LEFT(array_to_string(ARRAY_AGG(DISTINCT(initcap(category_identifier))), ', ') ,57) || '...' 
      ELSE array_to_string(ARRAY_AGG(DISTINCT(initcap(category_identifier))), ', ')
    END AS categories,
    mode() WITHIN GROUP (ORDER BY situation_on) AS situation_on, '' AS comment, mode() WITHIN GROUP (ORDER BY rate) AS rate
  FROM tedb.v_vat_rate tedb_vr
  JOIN latest ON tedb_vr.member_state = latest.member_state AND tedb_vr.type = latest.type AND tedb_vr.category_fk = latest.category_fk
  WHERE tedb_vr.type = 'REDUCED' AND situation_on = latest.max_situation_on
  GROUP BY 1,2;