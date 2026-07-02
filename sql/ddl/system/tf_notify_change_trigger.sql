
CREATE OR REPLACE FUNCTION system.notify_change_trigger()
  RETURNS trigger AS
$BODY$
DECLARE
  payload TEXT;
BEGIN
  -- exit if the trigger is fired by lysref_lis to stop reflis internal processing from generating update events
  IF current_user = 'lysref_lis' THEN
    RETURN null;
  END IF;

  payload := json_build_object('timestamp', NOW(), 'action', LOWER(TG_OP), 'schema', TG_TABLE_SCHEMA, 'table', TG_TABLE_NAME);
  PERFORM pg_notify('change', payload);
  RETURN null;
END;
$BODY$
LANGUAGE plpgsql VOLATILE
COST 100;
