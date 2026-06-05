
TRUNCATE TABLE process.step_link;
TRUNCATE TABLE process.point;
DELETE FROM process.step;
DELETE FROM process.run;
DELETE FROM process.flow;

ALTER SEQUENCE process.step_id_seq RESTART WITH 1;
ALTER SEQUENCE process.run_id_seq RESTART WITH 1;
ALTER SEQUENCE process.flow_id_seq RESTART WITH 1;