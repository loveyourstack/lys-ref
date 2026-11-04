
-- for smoke test
INSERT INTO tedb.vat_category (description, identifier) VALUES ('All VAT', 'All');

INSERT INTO tedb.vat_rate (
  category_fk,
  cn_codes,
  comment,
  cpa_codes,
  member_state,
  rate_type,
  rate,
  situation_on,
  "type"
) VALUES (1, ARRAY['CN123'], 'Test comment', ARRAY['CPA123'], 'DE', 'DEFAULT', 0.2, '2024-06-01', 'STANDARD');