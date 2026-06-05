
CREATE OR REPLACE FUNCTION derive_unique(_schema text, _table text, _tar_col text, _id bigint, VARIADIC _parts text[]) RETURNS text AS
$BODY$
DECLARE
	v_ret text = '';
	v_part text;
	v_new_arr text[];
	v_exists bool;
BEGIN

FOREACH v_part IN ARRAY _parts LOOP
	IF v_part IS NOT NULL AND v_part != '' THEN
		v_new_arr = array_append(v_new_arr, v_part);
	END IF;
END LOOP;

-- get name derived from parts
v_ret = array_to_string(v_new_arr, ' ');

-- see if the exact match already exists
EXECUTE format('SELECT EXISTS(SELECT 1 FROM %I.%I WHERE id != $1 AND %I = $2);', _schema, _table, _tar_col) USING _id, v_ret INTO v_exists;

-- if not found, this is unique and can return
IF NOT v_exists THEN
	RETURN v_ret;
END IF;

-- else return with new ID appended
RETURN v_ret || ' (ID '||_id||')';

END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = public, pg_temp
  COST 100;

/*
select * from hr.employee
select * from derive_unique('hr', 'employee', 'given_name', 10, 'Michaela')
*/