
CREATE TRIGGER t_notification_insert AFTER INSERT ON system.notification FOR EACH ROW EXECUTE PROCEDURE system.notification_trigger();