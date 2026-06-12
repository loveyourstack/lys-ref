
CREATE OR REPLACE FUNCTION system.notification_trigger()
  RETURNS trigger AS
$BODY$
BEGIN
  PERFORM pg_notify('system.notification', NEW.id::text);
  RETURN null;
END;
$BODY$
LANGUAGE plpgsql VOLATILE
COST 100;

