
CREATE OR REPLACE VIEW publisher.v_author AS
  SELECT
    pub_a.id,
    pub_a.created_at,
    pub_a.created_by,
    pub_a.last_user_update_by,
    pub_a.name,
    pub_a.updated_at,
    COALESCE(pub_b.cnt, 0) AS book_count
  FROM publisher.author pub_a
  LEFT JOIN (SELECT author_fk, count(*) AS cnt FROM publisher.book GROUP BY author_fk) pub_b ON pub_a.id = pub_b.author_fk;

CREATE OR REPLACE VIEW publisher.v_author_archived AS
  SELECT
    pub_a_arc.id,
    pub_a_arc.archived_at,
    pub_a_arc.archived_by_cascade,
    pub_a_arc.created_at,
    pub_a_arc.created_by,
    pub_a_arc.last_user_update_by,
    pub_a_arc.name,
    pub_a_arc.updated_at,
    COALESCE(pub_b_arc.cnt, 0) AS book_count
  FROM publisher.author_archived pub_a_arc
  LEFT JOIN (SELECT author_fk, count(*) AS cnt FROM publisher.book_archived GROUP BY author_fk) pub_b_arc ON pub_a_arc.id = pub_b_arc.author_fk;


-----------------------------------------------------------------------------------------------

CREATE OR REPLACE VIEW publisher.v_book AS
  SELECT
    pub_b.id,
    pub_b.author_fk,
      pub_a.name AS author,
    pub_b.created_at,
    pub_a.created_by,
    pub_b.last_user_update_by,
    pub_b.name,
    pub_b.updated_at
  FROM publisher.book pub_b
  JOIN publisher.author pub_a ON pub_b.author_fk = pub_a.id;

CREATE OR REPLACE VIEW publisher.v_book_archived AS
  SELECT
    pub_b_arc.id,
    pub_b_arc.archived_at,
    pub_b_arc.archived_by_cascade,
    pub_b_arc.author_fk,
      COALESCE(pub_a_arc.name,pub_a.name,'Unknown') AS author, -- check in both the archived and active author tables to get the author name
	    CASE WHEN pub_a.name IS NULL THEN true ELSE false END AS author_is_archived,
    pub_b_arc.created_at,
    pub_b_arc.created_by,
    pub_b_arc.last_user_update_by,
    pub_b_arc.name,
    pub_b_arc.updated_at
  FROM publisher.book_archived pub_b_arc
  LEFT JOIN publisher.author_archived pub_a_arc ON pub_b_arc.author_fk = pub_a_arc.id
  LEFT JOIN publisher.author pub_a ON pub_b_arc.author_fk = pub_a.id;
