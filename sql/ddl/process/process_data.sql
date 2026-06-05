
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

INSERT INTO process.step_link (step_fk, depends_on_fk) VALUES 
  ((SELECT id FROM process.step WHERE name = 'Aggregate costs'), (SELECT id FROM process.step WHERE name = 'Fetch XRates source')),
  ((SELECT id FROM process.step WHERE name = 'Aggregate costs'), (SELECT id FROM process.step WHERE name = 'Fetch cost source 1')),
  ((SELECT id FROM process.step WHERE name = 'Aggregate costs'), (SELECT id FROM process.step WHERE name = 'Fetch cost source 2')),
  ((SELECT id FROM process.step WHERE name = 'Aggregate costs'), (SELECT id FROM process.step WHERE name = 'Fetch cost source 3')),

  ((SELECT id FROM process.step WHERE name = 'Aggregate revenue'), (SELECT id FROM process.step WHERE name = 'Fetch XRates source')),
  ((SELECT id FROM process.step WHERE name = 'Aggregate revenue'), (SELECT id FROM process.step WHERE name = 'Fetch revenue source 1')),
  ((SELECT id FROM process.step WHERE name = 'Aggregate revenue'), (SELECT id FROM process.step WHERE name = 'Fetch revenue source 2')),

  ((SELECT id FROM process.step WHERE name = 'Create report'), (SELECT id FROM process.step WHERE name = 'Aggregate costs')),
  ((SELECT id FROM process.step WHERE name = 'Create report'), (SELECT id FROM process.step WHERE name = 'Aggregate revenue'))
;

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

INSERT INTO process.step_link (step_fk, depends_on_fk) VALUES 
  ((SELECT id FROM process.step WHERE name = 'Generate invoice PDFs'), (SELECT id FROM process.step WHERE name = 'Create invoice data')),
  ((SELECT id FROM process.step WHERE name = 'Send invoices'), (SELECT id FROM process.step WHERE name = 'Validate invoice recipients')),
  ((SELECT id FROM process.step WHERE name = 'Send invoices'), (SELECT id FROM process.step WHERE name = 'Generate invoice PDFs'))
;
