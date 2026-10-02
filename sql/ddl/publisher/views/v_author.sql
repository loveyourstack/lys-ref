
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
