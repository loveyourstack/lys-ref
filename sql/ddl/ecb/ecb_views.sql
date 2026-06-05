
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


CREATE OR REPLACE VIEW ecb.v_exchange_rate AS
  SELECT
    xr.id,
    xr.created_at,
    xr.day,
    xr.frequency,
    xr.from_currency_fk,
      from_curr.code AS from_currency,
    xr.rate,
    xr.to_currency_fk,
      to_curr.code AS to_currency,
    xr.updated_at
  FROM ecb.exchange_rate xr
  JOIN ecb.currency from_curr ON xr.from_currency_fk = from_curr.id
  JOIN ecb.currency to_curr ON xr.to_currency_fk = to_curr.id;


CREATE OR REPLACE VIEW ecb.v_xr_perf_uneven AS
  WITH counts AS (
    SELECT period, to_currency_code, count(*) AS cnt
    FROM ecb.xr_perf_normalized
    GROUP BY 1, 2
  )
  SELECT period, to_currency_code, cnt
  FROM (
    SELECT period, to_currency_code, cnt,
           min(cnt) OVER (PARTITION BY period) AS min_cnt,
           max(cnt) OVER (PARTITION BY period) AS max_cnt
    FROM counts
  ) t
  WHERE min_cnt != max_cnt
  ORDER BY 1, 2;
