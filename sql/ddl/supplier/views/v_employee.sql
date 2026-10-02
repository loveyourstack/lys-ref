
-- security_barrier -> ensures RLS policies are applied before any other function
-- security_invoker -> enables RLS policies and checks calling user's rights on all affected tables
CREATE OR REPLACE VIEW supplier.v_employee WITH (security_barrier, security_invoker) AS
  SELECT 
    s_e.id,
    s_e.company_fk,
      s_c.name AS company,
    s_e.created_at,
    s_e.email,
    s_e.family_name,
    s_e.given_name,
    s_e.name,
    s_e.updated_at
  FROM supplier.employee s_e
  JOIN supplier.company s_c ON s_e.company_fk = s_c.id;

GRANT SELECT ON supplier.v_employee TO lysref_supplier;

