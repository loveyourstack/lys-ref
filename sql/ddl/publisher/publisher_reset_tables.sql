
TRUNCATE TABLE publisher.book_archived;
TRUNCATE TABLE publisher.book;
TRUNCATE TABLE publisher.author_archived;
DELETE FROM publisher.author;

ALTER SEQUENCE publisher.author_id_seq RESTART WITH 1;
