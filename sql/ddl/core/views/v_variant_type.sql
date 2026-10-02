
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