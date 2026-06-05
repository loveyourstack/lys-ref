
INSERT INTO system.user (
  allow_multiple_sessions,
  created_by,
  email,
  email_verified,
  given_name,
  family_name,
  hashed_pw,
  name,
  roles
) VALUES (
  false,
  'Initialization',
  '<developer_email>',
  true,
  '<developer_givenName>',
  '<developer_familyName>',
  '<developer_hashedPw>',
  '<developer_givenName>',
  '{Tech}'::system.role[]
);

-- fake users for testing and development. not for use in production

INSERT INTO system.user (
  allow_multiple_sessions,
  created_by,
  email,
  email_verified,
  given_name,
  family_name,
  hashed_pw,
  name,
  roles
) VALUES 
  (true,
  'Initialization',
  'friend@example.com',
  true,
  'Friend',
  'User',
  '$2a$10$qTCP4TN7eVvGXSp0MeGrMOiGUODBUHPnrOTAK./UBZBr0eD1lznRm', -- f
  'Friend',
  '{Standard}'::system.role[]
  ),
  (true,
  'Initialization',
  'guest@example.com',
  true,
  'Guest',
  'User',
  '$2a$10$PvfLDsafljxuYqnEh.84qufxeS3LmRBSUNogYc9EkQofdAD0Bm0gy', -- g
  'Guest',
  '{Viewer}'::system.role[]
  )
;
