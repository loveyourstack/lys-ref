
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


CREATE OR REPLACE VIEW supplier.v_product WITH (security_barrier, security_invoker) AS
  SELECT
    s_p.id,
    s_p.category_fk,
    s_p.created_by,
    s_p.last_user_update_by,
      s_pc.name AS category,
    s_p.company_fk,
      s_c.name AS company,
    s_p.created_at,
    s_p.name,
    s_p.units_on_order,
    s_p.updated_at
  FROM supplier.product s_p
  JOIN supplier.product_category s_pc ON s_p.category_fk = s_pc.id
  JOIN supplier.company s_c ON s_p.company_fk = s_c.id;

GRANT SELECT ON supplier.v_product TO lysref_supplier;
