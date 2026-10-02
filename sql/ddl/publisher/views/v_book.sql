
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
