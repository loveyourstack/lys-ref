
CREATE OR REPLACE VIEW core.v_mandatory_value AS
  SELECT
    mv.id,
    mv.c_bool,
    mv.c_date_cet,
    mv.c_enum,
    mv.c_int,
    mv.c_numeric,
    mv.c_table_fk,
      geo_o.name AS c_table, -- indent joined columns
    mv.c_text,
    mv.c_time,
    mv.created_at,
    mv.updated_at
  FROM core.mandatory_value mv
  JOIN geo.ocean geo_o ON mv.c_table_fk = geo_o.id;


CREATE OR REPLACE VIEW core.v_optional_value AS
  SELECT
    ov.id,
    ov.c_bool,
    ov.c_date_cet,
    ov.c_enum,
    ov.c_int,
    ov.c_numeric,
    ov.c_table_fk,
      geo_c.name AS c_table,
    ov.c_text,
    ov.c_time,
    ov.created_at,
    ov.updated_at
  FROM core.optional_value ov
  JOIN geo.country geo_c ON ov.c_table_fk = geo_c.id;


CREATE OR REPLACE VIEW core.v_variant_type AS
  SELECT
    vt.id,
    vt.c_constrained_text,
    vt.c_ip,
    vt.c_long_text,
    CASE WHEN LENGTH(vt.c_long_text) > 47 THEN LEFT(vt.c_long_text, 47) || '...' ELSE vt.c_long_text END AS c_long_text_short,
    vt.c_money_amount,
    vt.c_percent,
    vt.created_at,
    vt.updated_at
  FROM core.variant_type vt;