
INSERT INTO process.flow (name, params) VALUES ('Create daily performance report', '{"fromDate={today-1}", "untilDate={today}"}');

INSERT INTO process.step (flow_fk, name, cmd, display_order) VALUES 
  (currval('process.flow_id_seq'), 'Fetch cost source 1', 'costsource1 :fromDate :untilDate', 110),
  (currval('process.flow_id_seq'), 'Fetch cost source 2', 'costsource2', 120),
  (currval('process.flow_id_seq'), 'Fetch cost source 3', 'costsource3', 130),

  (currval('process.flow_id_seq'), 'Fetch XRates source', 'xratessource Daily', 210),

  (currval('process.flow_id_seq'), 'Fetch revenue source 1', 'revsource1', 310),
  (currval('process.flow_id_seq'), 'Fetch revenue source 2', 'revsource2', 320),

  (currval('process.flow_id_seq'), 'Aggregate costs', 'aggcosts :fromDate :untilDate', 410),
  (currval('process.flow_id_seq'), 'Aggregate revenue', 'aggrevenue', 420),

  (currval('process.flow_id_seq'), 'Create report', 'createreports :fromDate :untilDate', 510)
;

WITH links AS (VALUES 
  ('Aggregate costs', 'Fetch XRates source'),
  ('Aggregate costs', 'Fetch cost source 1'),
  ('Aggregate costs', 'Fetch cost source 2'),
  ('Aggregate costs', 'Fetch cost source 3'),

  ('Aggregate revenue', 'Fetch XRates source'),
  ('Aggregate revenue', 'Fetch revenue source 1'),
  ('Aggregate revenue', 'Fetch revenue source 2'),

  ('Create report', 'Aggregate costs'),
  ('Create report', 'Aggregate revenue')
)
INSERT INTO process.step_link (step_fk, depends_on_fk)
  SELECT s.id, d.id
  FROM links l
  JOIN process.step s ON s.name = l.column1
  JOIN process.step d ON d.name = l.column2;


-- fake runs
INSERT INTO process.run (flow_fk, step_id, step_name) VALUES 
  (currval('process.flow_id_seq'), (SELECT id FROM process.step WHERE name = 'Fetch cost source 2'), 'Fetch cost source 2');

INSERT INTO process.point (cmd, depends_on, display_order, err_msg, finished_at, run_fk, started_at, status, step_id, step_name) VALUES 
  ('costsource2', '{}', 120, '', now() - INTERVAL '1 day' + INTERVAL '3 seconds', currval('process.run_id_seq'), now() - INTERVAL '1 day', 'Completed', (SELECT id FROM process.step WHERE name = 'Fetch cost source 2'), 'Fetch cost source 2')
;

INSERT INTO process.run (flow_fk, step_id, step_name) VALUES 
  (currval('process.flow_id_seq'), (SELECT id FROM process.step WHERE name = 'Aggregate revenue'), 'Aggregate revenue');

INSERT INTO process.point (cmd, depends_on, display_order, err_msg, finished_at, run_fk, started_at, status, step_id, step_name) VALUES 
  ('aggrevenue', '{}', 420, 'fake application error', now() - INTERVAL '1 day' + INTERVAL '1 seconds', currval('process.run_id_seq'), now() - INTERVAL '1 day', 'Error', (SELECT id FROM process.step WHERE name = 'Aggregate revenue'), 'Aggregate revenue')
;

------------------------------

INSERT INTO process.flow (name, params) VALUES ('Send monthly invoices', '{"month={current_month-1}"}');

INSERT INTO process.step (flow_fk, name, cmd, display_order) VALUES 
  (currval('process.flow_id_seq'), 'Create invoice data', 'invoices createdata :month', 110),

  (currval('process.flow_id_seq'), 'Generate invoice PDFs', 'invoices genpdfs :month', 210),

  (currval('process.flow_id_seq'), 'Validate invoice recipients', 'invoices valrecips', 310),

  (currval('process.flow_id_seq'), 'Send invoices', 'invoices send :month', 410)
;

WITH links AS (VALUES 
  ('Generate invoice PDFs', 'Create invoice data'),
  ('Send invoices', 'Validate invoice recipients'),
  ('Send invoices', 'Generate invoice PDFs')
)
INSERT INTO process.step_link (step_fk, depends_on_fk)
  SELECT s.id, d.id
  FROM links l
  JOIN process.step s ON s.name = l.column1
  JOIN process.step d ON d.name = l.column2;
