
TRUNCATE TABLE supplier.product;
TRUNCATE TABLE supplier.employee;
DELETE FROM supplier.product_category;
DELETE FROM supplier.company;

ALTER SEQUENCE supplier.product_category_id_seq RESTART WITH 1;
ALTER SEQUENCE supplier.company_id_seq RESTART WITH 1;
