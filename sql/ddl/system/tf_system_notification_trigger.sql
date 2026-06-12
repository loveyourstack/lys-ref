
CREATE OR REPLACE FUNCTION system.notification_trigger()
  RETURNS trigger AS
$BODY$
DECLARE
  payload TEXT;
BEGIN
  payload := json_build_object(
    'id', NEW.id,
    'type', NEW.not_type,
    'user_id', NEW.user_fk
  );
  PERFORM pg_notify('system.notification', payload);
  RETURN null;
END;
$BODY$
LANGUAGE plpgsql VOLATILE
COST 100;
