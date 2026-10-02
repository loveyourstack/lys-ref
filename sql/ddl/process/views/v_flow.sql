
CREATE OR REPLACE VIEW process.v_flow AS
  SELECT
    proc_f.id,
    proc_f.created_at,
    proc_f.name,
    proc_f.params,
    COALESCE(array_to_string(ARRAY(
      SELECT process.f_replace_dates(param)
      FROM unnest(COALESCE(proc_f.params, ARRAY[]::text[])) AS param
    ), ' '), '') AS params_replaced,
    proc_f.updated_at,
    COALESCE(proc_s.cnt,0) AS step_count,
    COALESCE(proc_r.cnt,0) AS run_count
  FROM process.flow proc_f
  LEFT JOIN (SELECT flow_fk, count(*) AS cnt FROM process.step GROUP BY 1) proc_s ON proc_s.flow_fk = proc_f.id
  LEFT JOIN (SELECT flow_fk, count(*) AS cnt FROM process.run GROUP BY 1) proc_r ON proc_r.flow_fk = proc_f.id;
