
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
