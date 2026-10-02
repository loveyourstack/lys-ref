
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
