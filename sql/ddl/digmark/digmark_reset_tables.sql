
TRUNCATE TABLE digmark.campaign_performance_aggregated;
TRUNCATE TABLE digmark.campaign_performance;
DELETE FROM digmark.campaign;
DELETE FROM digmark.launcher_fb;
DELETE FROM digmark.launcher_gads;
DELETE FROM digmark.vertical;

ALTER SEQUENCE digmark.campaign_id_seq RESTART WITH 1;
ALTER SEQUENCE digmark.launcher_seq RESTART WITH 1;
ALTER SEQUENCE digmark.vertical_id_seq RESTART WITH 1;