
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
