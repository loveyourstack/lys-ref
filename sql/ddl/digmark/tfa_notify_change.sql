
CREATE TRIGGER t_launcher_fb_notify_change AFTER INSERT OR UPDATE ON digmark.launcher_fb FOR EACH ROW EXECUTE PROCEDURE system.tf_notify_change();
CREATE TRIGGER t_launcher_gads_notify_change AFTER INSERT OR UPDATE ON digmark.launcher_gads FOR EACH ROW EXECUTE PROCEDURE system.tf_notify_change();
