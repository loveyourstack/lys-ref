
CREATE OR REPLACE VIEW process.v_flow AS
  SELECT
    proc_f.id,
    proc_f.created_at,
    proc_f.name,
    proc_f.params,
    COALESCE(array_to_string(ARRAY(
      SELECT process.replace_dates(param)
      FROM unnest(COALESCE(proc_f.params, ARRAY[]::text[])) AS param
    ), ' '), '') AS params_replaced,
    proc_f.updated_at,
    COALESCE(proc_s.cnt,0) AS step_count,
    COALESCE(proc_r.cnt,0) AS run_count
  FROM process.flow proc_f
  LEFT JOIN (SELECT flow_fk, count(*) AS cnt FROM process.step GROUP BY 1) proc_s ON proc_s.flow_fk = proc_f.id
  LEFT JOIN (SELECT flow_fk, count(*) AS cnt FROM process.run GROUP BY 1) proc_r ON proc_r.flow_fk = proc_f.id;


CREATE OR REPLACE VIEW process.v_step AS
  SELECT
    proc_s.id,
    proc_s.cmd,
    proc_s.created_at,
    proc_s.display_order,
    proc_s.flow_fk,
      proc_f.name AS flow,
    proc_s.name,
    proc_s.updated_at,
    COALESCE(proc_sl.deps, '{}') AS depends_on,
    COALESCE(proc_sl.dep_names, '{}') AS depends_on_names,
    COALESCE(proc_p.cnt,0) AS point_count
  FROM process.step proc_s
  JOIN process.flow proc_f ON proc_s.flow_fk = proc_f.id
  LEFT JOIN (
    SELECT step_fk, ARRAY_AGG(depends_on_fk ORDER BY depends_on_fk) AS deps, 
      ARRAY_AGG(sp.name || ' | ' || sl.id  ORDER BY sp.name) AS dep_names -- for use in UI: displaying dep, and the ID to delete
    FROM process.step_link sl JOIN process.step sp ON sl.depends_on_fk = sp.id GROUP BY 1
  ) proc_sl ON proc_sl.step_fk = proc_s.id
  LEFT JOIN (SELECT step_id, count(*) AS cnt FROM process.point GROUP BY 1) proc_p ON proc_p.step_id = proc_s.id;


CREATE OR REPLACE VIEW process.v_run AS
  SELECT
    proc_r.id,
    proc_r.created_at,
    proc_r.flow_fk,
      proc_f.name AS flow,
    proc_r.step_id,
    proc_r.step_name,
    COALESCE(proc_p.finished_at,'0001-01-01 12:00:00') AS finished_at,
    COALESCE(proc_p.point_count,0) AS point_count,
    COALESCE(proc_p.started_at,'0001-01-01 12:00:00') AS started_at,
    COALESCE(proc_p_stati.val,'') AS point_stati
  FROM process.run proc_r
  JOIN process.flow proc_f ON proc_r.flow_fk = proc_f.id
  LEFT JOIN (SELECT run_fk, min(started_at) FILTER (WHERE started_at > '1970-01-01') AS started_at, max(finished_at) AS finished_at, count(*) AS point_count FROM process.point GROUP BY 1) proc_p ON proc_p.run_fk = proc_r.id
  LEFT JOIN (
    WITH stati AS (
      SELECT run_fk, status, count(*) AS cnt FROM process.point GROUP BY 1,2
    )
    SELECT run_fk, array_to_string(array_agg(status || ': ' || cnt ORDER BY status), ', ') AS val FROM stati GROUP BY 1) proc_p_stati ON proc_p_stati.run_fk = proc_r.id;