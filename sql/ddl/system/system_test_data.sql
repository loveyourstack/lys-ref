
-- for smoke test
INSERT INTO system.blocked_ip (ip) VALUES ('1.1.1.1');

INSERT INTO system.ext_data_sync (last_sync_at, sync_key) VALUES ('2024-06-01', 'test_key');

-- system.login_attempt and system.session are not seeded here: the refsrv tests empty them, so their smoke test inserts its own rows

INSERT INTO system.notification (is_read, message, not_type, user_fk) VALUES (true, 'Test message', 'Info', 1);

INSERT INTO system.rpc_login_attempt (created_at, ip, is_blocked, num_attempts) VALUES (now(), '1.1.1.1', true, 5);

INSERT INTO system.rpc_request (duration_ms, endpoint, ip, status, user_name) VALUES (100, 'test_endpoint', '1.1.1.1', 'OK', 'test_user');

INSERT INTO system.rpc_session (
  allow_multiple_sessions,
  created_at,
  email,
  expires_at,
  family_name,
  force_password_change,
  given_name,
  geo_ip_country_iso_code,
  geo_ip_location,
  ip,
  last_access_at,
  profile_pic,
  roles,
  token,
  user_agent,
  user_id,
  user_name
) VALUES (true, now(), 'test@example.com', now() + interval '1 day', 'Doe', false, 'John', 'US', 'Test Location', '1.1.1.1', now(), 'pp', ARRAY['Viewer'::system.role], 'token', 'ua', 1, 'test_user');

INSERT INTO system.rpc_session_history (
  allow_multiple_sessions,
  created_at,
  email,
  expires_at,
  family_name,
  force_password_change,
  given_name,
  geo_ip_country_iso_code,
  geo_ip_location,
  ip,
  last_access_at,
  profile_pic,
  roles,
  token,
  user_agent,
  user_id,
  user_name,
  deleted_at
) VALUES (true, now(), 'test@example.com', now() + interval '1 day', 'Doe', false, 'John', 'US', 'Test Location', '1.1.1.1', now(), 'pp', ARRAY['Viewer'::system.role], 'token', 'ua', 1, 'test_user', now());

INSERT INTO system.session_history (
  allow_multiple_sessions,
  created_at,
  email,
  expires_at,
  family_name,
  force_password_change,
  given_name,
  geo_ip_country_iso_code,
  geo_ip_location,
  ip,
  last_access_at,
  profile_pic,
  roles,
  token,
  user_agent,
  user_fk,
  user_name,
  deleted_at
) VALUES (true, now(), 'test@example.com', now() + interval '1 day', 'Doe', false, 'John', 'US', 'Test Location', '1.1.1.1', now(), 'pp', ARRAY['Viewer'::system.role], 'token', 'ua', 1, 'test_user', now());

INSERT INTO system.srv_request (duration_ms, endpoint, ip, method, status_code, user_name) VALUES (100, 'test_endpoint', '1.1.1.1', 'GET', 200, 'test_user');