
-- authors
INSERT INTO lyspgmon.audit_update (affected_schema, affected_table, affected_id, affected_by, affected_at, affected_old_values, affected_new_values) VALUES 
  ('publisher', 'author', (SELECT id FROM publisher.author WHERE name = 'Ava Mercer'), 'Anna', now() - interval '36 hours', '{"name": "Ava Gardner"}', '{"name": "Ava Merce"}'),
  ('publisher', 'author', (SELECT id FROM publisher.author WHERE name = 'Ava Mercer'), 'Michael', now() - interval '12 hours', '{"name": "Ava Merce"}', '{"name": "Ava Mercer"}'),
  ('publisher', 'author', (SELECT id FROM publisher.author WHERE name = 'Benji Hart'), 'Paul', now() - interval '48 hours', '{"name": "Benjy Hart"}', '{"name": "Benji Hart"}')
;

-- books
INSERT INTO lyspgmon.audit_update (affected_schema, affected_table, affected_id, affected_by, affected_at, affected_old_values, affected_new_values) VALUES 
  ('publisher', 'book', (SELECT id FROM publisher.book WHERE name = 'Harbor of Small Days'), 'Anna', now() - interval '36 hours', '{"author_fk": 2, "name": "Harbor of Days"}', '{"author_fk": 3, "name": "Harbor of Small Day"}'),
  ('publisher', 'book', (SELECT id FROM publisher.book WHERE name = 'Harbor of Small Days'), 'Michael', now() - interval '12 hours', '{"name": "Harbor of Small Day"}', '{"name": "Harbor of Small Days"}'),
  ('publisher', 'book', (SELECT id FROM publisher.book WHERE name = 'Rain Over Alder Street'), 'Paul', now() - interval '48 hours', '{"name": "Water Over Alder Street"}', '{"name": "Rain Over Alder Street"}')
;