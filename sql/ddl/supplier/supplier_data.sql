
INSERT INTO supplier.product_category (name, description) VALUES
  ('Beverages','Soft drinks, coffees, teas, beers, and ales'),
  ('Condiments','Sweet and savory sauces, relishes, spreads, and seasonings'),
  ('Confections','Desserts, candies, and sweet breads'),
  ('Dairy Products','Cheeses'),
  ('Grains/Cereals','Breads, crackers, pasta, and cereal'),
  ('Meat/Poultry','Prepared meats'),
  ('Produce','Dried fruit and bean curd'),
  ('Seafood','Seaweed and fish');

-------------------------------------------------------------------------------------------

INSERT INTO supplier.company (name) VALUES ('G''day, Mate');

INSERT INTO supplier.employee (company_fk, email, given_name, family_name) VALUES 
  (currval('supplier.company_id_seq'), 'sarah.lucas@example.com', 'Sarah', 'Lucas'),
  (currval('supplier.company_id_seq'), 'wendy.mackenzie@example.com', 'Wendy', 'Mackenzie');

INSERT INTO supplier.product(category_fk, company_fk, name, units_on_order, created_by, last_user_update_by) VALUES 
  ((SELECT id FROM supplier.product_category WHERE name = 'Meat/Poultry'), currval('supplier.company_id_seq'), 'Perth Pasties', 50, 'Sarah Lucas', 'Wendy Mackenzie'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Produce'), currval('supplier.company_id_seq'), 'Manjimup Dried Apples', 20, 'Wendy Mackenzie', 'Wendy Mackenzie'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Grains/Cereals'), currval('supplier.company_id_seq'), 'Filo Mix', 0, 'Sarah Lucas', '');

------------------------------

INSERT INTO supplier.company (name) VALUES ('Formaggi Fortini s.r.l.');

INSERT INTO supplier.employee (company_fk, email, given_name, family_name) VALUES 
  (currval('supplier.company_id_seq'), 'elio.rossi@example.com', 'Elio', 'Rossi'),
  (currval('supplier.company_id_seq'), 'maria.bianchi@example.com', 'Maria', 'Bianchi'),
  (currval('supplier.company_id_seq'), 'giovanni.verdi@example.com', 'Giovanni', 'Verdi');

INSERT INTO supplier.product(category_fk, company_fk, name, units_on_order, created_by, last_user_update_by) VALUES 
  ((SELECT id FROM supplier.product_category WHERE name = 'Dairy Products'), currval('supplier.company_id_seq'), 'Gorgonzola Telino', 48, 'Elio Rossi', 'Elio Rossi'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Dairy Products'), currval('supplier.company_id_seq'), 'Mascarpone Fabioli', 30, 'Maria Bianchi', 'Maria Bianchi'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Dairy Products'), currval('supplier.company_id_seq'), 'Mozzarella di Giovanni', 5, 'Giovanni Verdi', 'Elio Rossi');

------------------------------

INSERT INTO supplier.company (name) VALUES ('Specialty Biscuits, Ltd.');

INSERT INTO supplier.employee (company_fk, email, given_name, family_name) VALUES 
  (currval('supplier.company_id_seq'), 'peter.wilson@example.com', 'Peter', 'Wilson'),
  (currval('supplier.company_id_seq'), 'susan.miller@example.com', 'Susan', 'Miller');

INSERT INTO supplier.product(category_fk, company_fk, name, units_on_order, created_by, last_user_update_by) VALUES 
  ((SELECT id FROM supplier.product_category WHERE name = 'Beverages'), currval('supplier.company_id_seq'), 'Chai', 1000, 'Peter Wilson', 'Peter Wilson'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Teatime Chocolate Biscuits', 2000, 'Susan Miller', 'Peter Wilson'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Sir Rodney''s Marmalade', 40, 'Susan Miller', 'Peter Wilson'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Sir Rodney''s Scones', 0, 'Susan Miller', ''),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Scottish Longbreads', 800, 'Peter Wilson', 'Susan Miller');

------------------------------

INSERT INTO supplier.company (name) VALUES ('Refrescos Americanas LTDA');

INSERT INTO supplier.employee (company_fk, email, given_name, family_name) VALUES 
  (currval('supplier.company_id_seq'), 'carlos.diaz@example.com', 'Carlos', 'Diaz');

INSERT INTO supplier.product(category_fk, company_fk, name, units_on_order, created_by, last_user_update_by) VALUES 
  ((SELECT id FROM supplier.product_category WHERE name = 'Beverages'), currval('supplier.company_id_seq'), 'Guaraná Fantástica', 1200, 'Carlos Diaz', 'Carlos Diaz');

------------------------------

INSERT INTO supplier.company (name) VALUES ('Heli Süßwaren GmbH & Co. KG');

INSERT INTO supplier.employee (company_fk, email, given_name, family_name) VALUES 
  (currval('supplier.company_id_seq'), 'petra.winkler@example.com', 'Petra', 'Winkler'),
  (currval('supplier.company_id_seq'), 'thomas.schmidt@example.com', 'Thomas', 'Schmidt'),
  (currval('supplier.company_id_seq'), 'lena.meyer@example.com', 'Lena', 'Meyer');

INSERT INTO supplier.product(category_fk, company_fk, name, units_on_order, created_by, last_user_update_by) VALUES 
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'NuNuCa Nuß-Nougat-Creme', 60, 'Petra Winkler', 'Thomas Schmidt'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Gumbär Gummibärchen', 400, 'Thomas Schmidt', 'Thomas Schmidt'),
  ((SELECT id FROM supplier.product_category WHERE name = 'Confections'), currval('supplier.company_id_seq'), 'Schoggi Schokolade', 80, 'Lena Meyer', 'Lena Meyer');

