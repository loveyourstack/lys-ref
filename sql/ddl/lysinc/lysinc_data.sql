
INSERT INTO lysinc.department (id, name) VALUES 
(1, 'Leadership'), (2, 'Engineering'), (3, 'Product');

------------------------------

-- profile pics were generated using https://thispersondoesnotexist.com/, resized to 400x400, then uploaded to S3 under /profiles

-- CEO: id 1
INSERT INTO lysinc.employee (id, department_fk, email, family_name, given_name, sex, job_title, profile_pic, reports_to, honorific, date_of_birth, join_date) VALUES
(1, 1, 'elena.ramirez@example.com', 'Ramirez', 'Elena', 'Female', 'Chief Executive Officer', 'elena_ramirez.jpg', 1, 'Mrs', '1978-04-12', '2014-06-16')
;

ALTER SEQUENCE lysinc.employee_id_seq RESTART WITH 2;

-- Engineering: initially reporting to CEO (id 1); first row (VP of Engineering) is the department head
INSERT INTO lysinc.employee (department_fk, email, family_name, given_name, sex, job_title, profile_pic, reports_to, honorific, date_of_birth, join_date) VALUES
(2, 'nathan.walsh@example.com', 'Walsh', 'Nathan', 'Male', 'VP of Engineering', 'nathan_walsh.jpg', 1, 'Mr', '1979-12-01', '2013-09-09'),
(2, 'priya.natarajan@example.com', 'Natarajan', 'Priya', 'Female', 'Engineering Manager', 'priya_natarajan.jpg', 1, 'Ms', '1986-09-23', '2018-02-12'),
(2, 'diego.fernandez@example.com', 'Fernandez', 'Diego', 'Male', 'Engineering Manager', 'diego_fernandez.jpg', 1, 'Mr', '1983-11-07', '2016-08-22'),
(2, 'sofia.kowalski@example.com', 'Kowalski', 'Sofia', 'Female', 'Senior Software Engineer', 'sofia_kowalski.jpg', 1, 'Mrs', '1990-01-18', '2019-04-15'),
(2, 'megan.foster@example.com', 'Foster', 'Megan', 'Female', 'Senior Software Engineer', 'megan_foster.jpg', 1, 'Ms', '1988-06-30', '2017-10-02'),
(2, 'liam.bryant@example.com', 'Bryant', 'Liam', 'Male', 'Senior Software Engineer', 'liam_bryant.jpg', 1, 'Mr', '1987-03-14', '2018-11-19'),
(2, 'claire.whitfield@example.com', 'Whitfield', 'Claire', 'Female', 'Senior Software Engineer', 'claire_whitfield.jpg', 1, 'Mrs', '1991-12-05', '2020-01-27'),
(2, 'hannah.schmidt@example.com', 'Schmidt', 'Hannah', 'Female', 'Software Engineer', 'hannah_schmidt.jpg', 1, 'Ms', '1995-05-21', '2022-03-07'),
(2, 'tomas.silva@example.com', 'Silva', 'Tomas', 'Male', 'Software Engineer', 'tomas_silva.jpg', 1, 'Mr', '1993-08-16', '2021-06-14'),
(2, 'fatima.hassan@example.com', 'Hassan', 'Fatima', 'Female', 'Software Engineer', 'fatima_hassan.jpg', 1, 'Mrs', '1996-10-28', '2023-01-09'),
(2, 'nikolai.petrov@example.com', 'Petrov', 'Nikolai', 'Male', 'Site Reliability Engineer', 'nikolai_petrov.jpg', 1, 'Mr', '1989-02-11', '2019-09-30'),
(2, 'emily.carter@example.com', 'Carter', 'Emily', 'Female', 'Site Reliability Engineer', 'emily_carter.jpg', 1, 'Ms', '1992-07-03', '2020-08-17'),
(2, 'owen.bennett@example.com', 'Bennett', 'Owen', 'Male', 'Site Reliability Engineer', 'owen_bennett.jpg', 1, 'Mr', '1994-04-26', '2022-09-12')
;

-- update Engineering reports_to

-- Engineering Managers report to VP of Engineering
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'VP of Engineering')
  WHERE department_fk = 2 AND job_title = 'Engineering Manager';

-- Software Engineers report to EM Priya
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'Engineering Manager' AND given_name = 'Priya')
  WHERE department_fk = 2 AND job_title IN ('Software Engineer', 'Senior Software Engineer');

-- Site Reliability Engineers report to EM Diego
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'Engineering Manager' AND given_name = 'Diego')
  WHERE department_fk = 2 AND job_title = 'Site Reliability Engineer';


-- Product: initially reporting to CEO (id 1); first row (VP of Product) is the department head
INSERT INTO lysinc.employee (department_fk, email, family_name, given_name, sex, job_title, profile_pic, reports_to, honorific, date_of_birth, join_date) VALUES
 (3, 'isabelle.laurent@example.com', 'Laurent', 'Isabelle', 'Female', 'VP of Product', 'isabelle_laurent.jpg', 1, 'Mrs', '1981-09-09', '2015-03-23'),
 (3, 'rahul.deshpande@example.com', 'Deshpande', 'Rahul', 'Male', 'Product Manager', 'rahul_deshpande.jpg', 1, 'Mr', '1989-06-17', '2019-02-04'),
 (3, 'chloe.dubois@example.com', 'Dubois', 'Chloe', 'Female', 'Product Manager', 'chloe_dubois.jpg', 1, 'Ms', '1991-10-02', '2020-05-11'),
 (3, 'andrea.vargas@example.com', 'Vargas', 'Andrea', 'Female', 'Product Designer', 'andrea_vargas.jpg', 1, 'Ms', '1993-01-25', '2021-09-06'),
 (3, 'rachel.turner@example.com', 'Turner', 'Rachel', 'Female', 'Product Analyst', 'rachel_turner.jpg', 1, 'Mrs', '1995-07-19', '2022-11-14'),
 (3, 'samuel.alves@example.com', 'Alves', 'Samuel', 'Male', 'Associate Product Manager', 'samuel_alves.jpg', 1, 'Mr', '1997-03-08', '2024-02-26')
;

-- update Product reports_to: all Product employees except the VP report to the VP of Product
UPDATE lysinc.employee SET reports_to = (SELECT id FROM lysinc.employee WHERE job_title = 'VP of Product')
  WHERE department_fk = 3 AND job_title != 'VP of Product';

------------------------------


