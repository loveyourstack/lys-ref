
CREATE OR REPLACE VIEW ecb.v_currency AS
  SELECT
    curr.id,
    curr.code,
    curr.created_at,
    curr.name,
    curr.updated_at,
    COALESCE(curr_md.id, 0) AS metadata_id,
    COALESCE(curr_md.is_active, false) AS is_active,
    COALESCE(curr_md.symbol, '') AS symbol
  FROM ecb.currency curr
  LEFT JOIN ecb.currency_metadata curr_md ON curr.code = curr_md.code;
