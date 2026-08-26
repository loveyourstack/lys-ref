
CREATE OR REPLACE FUNCTION lysinc.events_in_month (
	_date date
)
RETURNS TABLE (
  department text,
  employee_id bigint,
  event_date date,
  event_type lysinc.event_type,
  full_name text,
  job_title text,
  sex lysinc.sex,
  years int
) AS
$BODY$

SELECT
  l_d.name AS department,
  l_e.id AS employee_id,
  make_date(EXTRACT(YEAR FROM _date)::int, EXTRACT(MONTH FROM l_e.date_of_birth)::int, EXTRACT(DAY FROM l_e.date_of_birth)::int) AS event_date,
  'Birthday'::lysinc.event_type AS event_type,
  l_e.full_name,
  l_e.job_title,
  l_e.sex,
  0 AS years -- discretion
FROM lysinc.employee l_e
JOIN lysinc.department l_d ON l_e.department_fk = l_d.id
WHERE EXTRACT(MONTH FROM l_e.date_of_birth) = EXTRACT(MONTH FROM _date)

UNION ALL

SELECT
  l_d.name AS department,
  l_e.id AS employee_id,
  make_date(EXTRACT(YEAR FROM _date)::int, EXTRACT(MONTH FROM l_e.join_date)::int, EXTRACT(DAY FROM l_e.join_date)::int) AS event_date,
  'Anniversary'::lysinc.event_type AS event_type,
  l_e.full_name,
  l_e.job_title,
  l_e.sex,
  EXTRACT(YEAR FROM _date)::int - EXTRACT(YEAR FROM l_e.join_date)::int AS years
FROM lysinc.employee l_e
JOIN lysinc.department l_d ON l_e.department_fk = l_d.id
WHERE EXTRACT(MONTH FROM l_e.join_date) = EXTRACT(MONTH FROM _date);

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;