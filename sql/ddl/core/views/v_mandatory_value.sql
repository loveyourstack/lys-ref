
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
