
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
