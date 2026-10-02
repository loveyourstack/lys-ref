
CREATE VIEW lysinc.v_employee AS
  SELECT 
    l_e.id,
    l_e.created_at,
    l_e.date_of_birth,
    l_e.department_fk,
      l_d.name AS department,
    l_e.email,
    l_e.family_name,
    l_e.full_name,
    l_e.given_name,
    l_e.honorific,
    l_e.job_title,
    l_e.join_date,
    l_e.profile_pic,
    l_e.reports_to,
      CASE WHEN l_e.id = l_e_rep.id THEN '' ELSE l_e_rep.full_name END AS reports_to_full_name,
      CASE WHEN l_e.id = l_e_rep.id THEN '' ELSE l_e_rep.job_title END AS reports_to_job_title,
    l_e.sex,
    l_e.updated_at
  FROM lysinc.employee l_e
  JOIN lysinc.employee l_e_rep ON l_e.reports_to = l_e_rep.id
  JOIN lysinc.department l_d ON l_e.department_fk = l_d.id;