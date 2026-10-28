
-- development service account
INSERT INTO system.service_account (
  api_key_hash,
  expires_at,
  is_active,
  last_access_at,
  name,
  roles
) VALUES (
  '186ef2a547a06bb9d275a5912eb7cf90b9c7b27a1bbd7349e00b1f109b4eb6cb', -- WvyRcwbTRUSnprjcLllzpKCazXnVbapH / dev only: replace in prod
  '9999-12-31 12:00:00+00',
  true,
  '0001-01-01 12:00:00+00',
  'RefSrv Monitor',
  '{Monitor}'::system.role[]
);


-- dev user
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
