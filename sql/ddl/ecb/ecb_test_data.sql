
-- for smoke test
INSERT INTO ecb.currency (code, name) VALUES
  ('USD', 'United States Dollar'),
  ('EUR', 'Euro')
;

INSERT INTO ecb.xr_perf_normalized (
  day,
  from_currency_fk,
  from_currency_code,
  normalized_perf,
  period,
  rate,
  to_currency_fk,
  to_currency_code
) VALUES ('2024-06-01', 1, 'USD', 1.2345, 'Today', 1.2345, 2, 'EUR');