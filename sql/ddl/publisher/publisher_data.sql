
INSERT INTO publisher.author (name) VALUES 
  ('Zara Okafor'),
  ('Benji Hart');


INSERT INTO publisher.author (name) VALUES ('Ava Mercer');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'Harbor of Small Days'),
  (currval('publisher.author_id_seq'), 'The Last Paper Atlas'),
  (currval('publisher.author_id_seq'), 'Rain Over Alder Street');

INSERT INTO publisher.author (name) VALUES ('Jonah Pike');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'A Season for Glass');

INSERT INTO publisher.author (name) VALUES ('Elena Corbett');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Library of Dusk'),
  (currval('publisher.author_id_seq'), 'Echoes from the Greenhouse'),
  (currval('publisher.author_id_seq'), 'Salt House Letters'),
  (currval('publisher.author_id_seq'), 'Winter in Signal Bay');

INSERT INTO publisher.author (name) VALUES ('Daniel Hsu');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'Monsoon Ledger'),
  (currval('publisher.author_id_seq'), 'The Seventh Orchard'),
  (currval('publisher.author_id_seq'), 'A Table by the Window');

INSERT INTO publisher.author (name) VALUES ('Sofia Maren');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Lantern District'),
  (currval('publisher.author_id_seq'), 'Blue Hour Dispatches'),
  (currval('publisher.author_id_seq'), 'The Cartographer''s Porch');

INSERT INTO publisher.author (name) VALUES ('Mila Duarte');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Garden of Lost Hours'),
  (currval('publisher.author_id_seq'), 'A Map for the Weary'),
  (currval('publisher.author_id_seq'), 'The Northbound Sketchbook'),
  (currval('publisher.author_id_seq'), 'Stone Fruit Summer'),
  (currval('publisher.author_id_seq'), 'The Long Evening Train');

INSERT INTO publisher.author (name) VALUES ('Theo Bennett');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'Borrowed Light');

INSERT INTO publisher.author (name) VALUES ('Hannah Sloane');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'Ashes in the Reading Room'),
  (currval('publisher.author_id_seq'), 'A Brief History of Rainfall');

INSERT INTO publisher.author (name) VALUES ('Owen Castillo');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Olive Station'),
  (currval('publisher.author_id_seq'), 'Night Market Geometry');

INSERT INTO publisher.author (name) VALUES ('Leila Haddad');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'A City of Small Fireworks');

INSERT INTO publisher.author (name) VALUES ('Noah Whitaker');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Copper Notebook'),
  (currval('publisher.author_id_seq'), 'Tides of Amber'),
  (currval('publisher.author_id_seq'), 'June After Midnight');

INSERT INTO publisher.author (name) VALUES ('Camila Estrada');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Windowmaker'),
  (currval('publisher.author_id_seq'), 'Fallow Season'),
  (currval('publisher.author_id_seq'), 'The Orchard of Forgotten Stars'),
  (currval('publisher.author_id_seq'), 'A Thousand Rooftops'),
  (currval('publisher.author_id_seq'), 'The Empty Conservatory');

INSERT INTO publisher.author (name) VALUES ('Nora Gallagher');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'Field Notes from Larkspur'),
  (currval('publisher.author_id_seq'), 'Dust and Evergreen');

INSERT INTO publisher.author (name) VALUES ('Julian Park');
INSERT INTO publisher.book (author_fk, name) VALUES
  (currval('publisher.author_id_seq'), 'The Quiet Frequency'),
  (currval('publisher.author_id_seq'), 'Letters from Meridian'),
  (currval('publisher.author_id_seq'), 'Paper Moons of August'),
  (currval('publisher.author_id_seq'), 'The Orchard Telegraph'),
  (currval('publisher.author_id_seq'), 'The House on Bellweather Lane'),
  (currval('publisher.author_id_seq'), 'Morning Train to Calder');


------------------------------------------------------------------------------------------------------

INSERT INTO publisher.author_archived (id, name, archived_at, archived_by_cascade) VALUES 
  (nextval('publisher.author_id_seq'), 'Marcus Vale', '2026-01-15 10:00:00', false),
  (nextval('publisher.author_id_seq'), 'Lila Chen', '2024-02-20 14:30:00', false),
  (nextval('publisher.author_id_seq'), 'Priya Nair', '2024-03-10 09:15:00', false),
  (nextval('publisher.author_id_seq'), 'Isaac Rowan', '2025-04-05 11:45:00', false),
  (nextval('publisher.author_id_seq'), 'Ethan Brooks', '2025-05-12 16:20:00', false);

-- cascaded book archive due to author archive
INSERT INTO publisher.book_archived (id, author_fk, name, archived_at, archived_by_cascade) VALUES
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author_archived WHERE name = 'Marcus Vale'), 'The Last Ember', '2026-01-15 10:00:00', true),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author_archived WHERE name = 'Marcus Vale'), 'Whispers in the Bamboo', '2024-02-20 14:30:00', true),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author_archived WHERE name = 'Priya Nair'), 'The Clockmaker''s Daughter', '2024-03-10 09:15:00', true),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author_archived WHERE name = 'Isaac Rowan'), 'Shadows of the Forgotten', '2025-04-05 11:45:00', true),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author_archived WHERE name = 'Ethan Brooks'), 'The Silent Harbor', '2025-05-12 16:20:00', true);

-- non-cascaded book archive (author still active)
INSERT INTO publisher.book_archived (id, author_fk, name, archived_at, archived_by_cascade) VALUES
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author WHERE name = 'Zara Okafor'), 'Giles and Fred - An Adventure', '2025-06-01 12:00:00', false),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author WHERE name = 'Zara Okafor'), 'The Rush to Dawn', '2025-07-15 09:30:00', false),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author WHERE name = 'Nora Gallagher'), 'Shadows True & Deep', '2025-08-20 14:45:00', false),
  (nextval('publisher.book_id_seq'), (SELECT id FROM publisher.author WHERE name = 'Benji Hart'), 'Your Time, My Time', '2025-09-10 11:00:00', false);


UPDATE publisher.author SET created_by = 'Initialization';
UPDATE publisher.author_archived SET created_by = 'Initialization';
UPDATE publisher.book SET created_by = 'Initialization';
UPDATE publisher.book_archived SET created_by = 'Initialization';