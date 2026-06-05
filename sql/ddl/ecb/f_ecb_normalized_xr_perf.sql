DROP FUNCTION IF EXISTS ecb.normalized_xr_perf;
CREATE OR REPLACE FUNCTION ecb.normalized_xr_perf (
  _from_curr_fk bigint,
  _days_before int,
	_days_after int
)
RETURNS TABLE (
  perf_day date,
  to_currency_fk bigint,
  to_currency_code varchar(3),
  rate numeric(12,4),
  normalized_perf numeric(8,4)
) AS
$BODY$
DECLARE
  v_start_day date;
  v_end_day date;
BEGIN

SELECT min(day) INTO v_start_day FROM ecb.exchange_rate WHERE from_currency_fk = _from_curr_fk AND day >= current_date + _days_before;
SELECT max(day) INTO v_end_day FROM ecb.exchange_rate WHERE from_currency_fk = _from_curr_fk AND day >= current_date + _days_before AND day <= current_date + _days_after;

RETURN QUERY

WITH active_curr AS (
  SELECT curr.id, curr.code
  FROM ecb.currency curr
  LEFT JOIN ecb.currency_metadata curr_md ON curr.code = curr_md.code
  WHERE COALESCE(curr_md.is_active, false) = true
),

base_rates AS (
  SELECT xr.day, xr.to_currency_fk, xr.rate, ac.code AS to_currency_code
  FROM ecb.exchange_rate xr
  JOIN active_curr ac ON ac.id = xr.to_currency_fk
  WHERE xr.frequency = 'D'
  AND xr.from_currency_fk = _from_curr_fk
  AND xr.day = v_start_day
)

SELECT
  xr.day,
  xr.to_currency_fk,
  br.to_currency_code,
  xr.rate,
  CASE WHEN br.rate = 0 THEN 0 ELSE (xr.rate / br.rate)::numeric(8,4) END AS normalized_perf
FROM ecb.exchange_rate xr
JOIN base_rates br USING (to_currency_fk)
WHERE xr.day >= v_start_day AND xr.day <= v_end_day
ORDER BY xr.day, br.to_currency_code;

END;
$BODY$
  LANGUAGE plpgsql STABLE
  SECURITY DEFINER
  SET search_path = ecb, pg_temp
  COST 100;
