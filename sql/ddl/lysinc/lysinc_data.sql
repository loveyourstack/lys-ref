
-- profile pics were generated using https://thispersondoesnotexist.com/, resized to 400x400, then uploaded to S3 under /profiles

-- CEO: id 1
INSERT INTO lysinc.employee (id, department, email, family_name, given_name, job_title, profile_pic, reports_to, honorific) VALUES
(1, 'Leadership', 'elena.ramirez@example.com', 'Ramirez', 'Elena', 'Chief Executive Officer', 'elena_ramirez.jpg', 1, 'Mrs')
;

ALTER SEQUENCE lysinc.employee_id_seq RESTART WITH 2;

-- Engineering: initially reporting to CEO (id 1); first row (VP of Engineering) is the department head
INSERT INTO lysinc.employee (department, email, family_name, given_name, job_title, profile_pic, reports_to, honorific) VALUES
('Engineering', 'nathan.walsh@example.com', 'Walsh', 'Nathan', 'VP of Engineering', 'nathan_walsh.jpg', 1, 'Mr'),
('Engineering', 'priya.natarajan@example.com', 'Natarajan', 'Priya', 'Engineering Manager', 'priya_natarajan.jpg', 1, 'Ms'),
('Engineering', 'diego.fernandez@example.com', 'Fernandez', 'Diego', 'Engineering Manager', 'diego_fernandez.jpg', 1, 'Mr'),
('Engineering', 'sofia.kowalski@example.com', 'Kowalski', 'Sofia', 'Senior Software Engineer', 'sofia_kowalski.jpg', 1, 'Mrs'),
('Engineering', 'megan.foster@example.com', 'Foster', 'Megan', 'Senior Software Engineer', 'megan_foster.jpg', 1, 'Ms'),
('Engineering', 'liam.bryant@example.com', 'Bryant', 'Liam', 'Senior Software Engineer', 'liam_bryant.jpg', 1, 'Mr'),
('Engineering', 'claire.whitfield@example.com', 'Whitfield', 'Claire', 'Senior Software Engineer', 'claire_whitfield.jpg', 1, 'Mrs'),
('Engineering', 'hannah.schmidt@example.com', 'Schmidt', 'Hannah', 'Software Engineer', 'hannah_schmidt.jpg', 1, 'Ms'),
('Engineering', 'tomas.silva@example.com', 'Silva', 'Tomas', 'Software Engineer', 'tomas_silva.jpg', 1, 'Mr'),
('Engineering', 'fatima.hassan@example.com', 'Hassan', 'Fatima', 'Software Engineer', 'fatima_hassan.jpg', 1, 'Mrs'),
('Engineering', 'nikolai.petrov@example.com', 'Petrov', 'Nikolai', 'Site Reliability Engineer', 'nikolai_petrov.jpg', 1, 'Mr'),
('Engineering', 'emily.carter@example.com', 'Carter', 'Emily', 'Site Reliability Engineer', 'emily_carter.jpg', 1, 'Ms'),
('Engineering', 'owen.bennett@example.com', 'Bennett', 'Owen', 'Site Reliability Engineer', 'owen_bennett.jpg', 1, 'Mr')
;

-- update Engineering reports_to

-- Engineering Managers report to VP of Engineering
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'VP of Engineering')
  WHERE department = 'Engineering' AND job_title = 'Engineering Manager';

-- Software Engineers report to EM Priya
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'Engineering Manager' AND given_name = 'Priya')
  WHERE department = 'Engineering' AND job_title IN ('Software Engineer', 'Senior Software Engineer');

-- Site Reliability Engineers report to EM Diego
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'Engineering Manager' AND given_name = 'Diego')
  WHERE department = 'Engineering' AND job_title = 'Site Reliability Engineer';


-- Product: initially reporting to CEO (id 1); first row (VP of Product) is the department head
INSERT INTO lysinc.employee (department, email, family_name, given_name, job_title, profile_pic, reports_to, honorific) VALUES
('Product', 'isabelle.laurent@example.com', 'Laurent', 'Isabelle', 'VP of Product', 'isabelle_laurent.jpg', 1, 'Mrs'),
('Product', 'rahul.deshpande@example.com', 'Deshpande', 'Rahul', 'Product Manager', 'rahul_deshpande.jpg', 1, 'Mr'),
('Product', 'chloe.dubois@example.com', 'Dubois', 'Chloe', 'Product Manager', 'chloe_dubois.jpg', 1, 'Ms'),
('Product', 'andrea.vargas@example.com', 'Vargas', 'Andrea', 'Product Designer', 'andrea_vargas.jpg', 1, 'Ms'),
('Product', 'rachel.turner@example.com', 'Turner', 'Rachel', 'Product Analyst', 'rachel_turner.jpg', 1, 'Mrs'),
('Product', 'samuel.alves@example.com', 'Alves', 'Samuel', 'Associate Product Manager', 'samuel_alves.jpg', 1, 'Mr')
;

-- update Product reports_to: all Product employees except the VP report to the VP of Product
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'VP of Product')
  WHERE department = 'Product' AND job_title != 'VP of Product';