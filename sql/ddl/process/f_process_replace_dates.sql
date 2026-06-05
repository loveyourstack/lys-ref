DROP FUNCTION IF EXISTS process.replace_dates;
CREATE OR REPLACE FUNCTION process.replace_dates(_input text) RETURNS text AS
$BODY$
DECLARE
	v_ret text;
	v_match text[];
	v_token text;
	v_offset integer;
BEGIN
	v_ret = _input;

  -- today without +/- modifier
	v_ret = replace(v_ret, '{today}', to_char(current_date, 'YYYY-MM-DD'));

  -- today with +/- modifier
	FOR v_match IN
		SELECT regexp_matches(v_ret, '\{today([+-][0-9]+)\}', 'g')
	LOOP
		v_token = '{today' || v_match[1] || '}';
		v_offset = v_match[1]::integer;
		v_ret = replace(v_ret, v_token, to_char(current_date + v_offset, 'YYYY-MM-DD'));
	END LOOP;

  -- current_month without +/- modifier
	v_ret = replace(v_ret, '{current_month}', to_char(date_trunc('month', current_date), 'YYYY-MM'));

  -- current_month with +/- modifier
	FOR v_match IN
		SELECT regexp_matches(v_ret, '\{current_month([+-][0-9]+)\}', 'g')
	LOOP
		v_token = '{current_month' || v_match[1] || '}';
		v_offset = v_match[1]::integer;
		v_ret = replace(v_ret, v_token, to_char(date_trunc('month', current_date) + make_interval(months => v_offset), 'YYYY-MM'));
	END LOOP;

	RETURN v_ret;
END;
$BODY$
  LANGUAGE plpgsql STABLE
  SECURITY DEFINER
  SET search_path = process, pg_temp
  COST 100;
