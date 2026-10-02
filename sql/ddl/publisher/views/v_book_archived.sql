
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
