
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
