/*
  Sample migration file for testing the replacement of DDL asset names.
*/

DROP FUNCTION digmark.f_campaign_optimizer;

DROP MATERIALIZED VIEW digmark.mv_latest_perf_summary;

DROP VIEW tedb.v_vat_rate_summary;
DROP VIEW tedb.v_vat_rate;

/* .. migration change would be here .. */

-- the following lines will be replaced with the corresponding DDL from the embedded assets

-- + tedb.v_vat_rate;
-- + tedb.v_vat_rate_summary;

-- + digmark.mv_latest_perf_summary;

-- + digmark.f_campaign_optimizer;
