-- name: SchemaReady
SELECT jsonb_build_object('ready', 1) FROM admins LIMIT 1;

-- name: LoginAdminByIdentifier
-- params: identifier text
SELECT jsonb_build_object(
    'id', id,
    'username', username,
    'email', email,
    'password_hash', password_hash,
    'full_name', full_name,
    'role', role,
    'is_active', is_active
) FROM admins
WHERE (lower(btrim(email)) = lower(btrim($1)) OR lower(btrim(username)) = lower(btrim($1)));

-- name: LoginAdminLock
-- params: admin_id bigint
SELECT jsonb_build_object(
    'id', id,
    'username', username,
    'email', email,
    'role', role,
    'is_active', is_active
) FROM admins
WHERE id = $1 FOR UPDATE;

-- name: LoginAdminUpdateLastLogin
-- params: admin_id bigint
UPDATE admins SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: LoginAdminSaveSession
-- params: admin_id bigint, device_id uuid, access_token_hash text, refresh_token_hash text, user_agent text, ip_address text, access_expires_at timestamptz, refresh_expires_at timestamptz
INSERT INTO admin_sessions (
    admin_id, device_id, access_token_hash, refresh_token_hash,
    user_agent, ip_address, access_expires_at, refresh_expires_at
) VALUES (
    $1, $2::uuid, $3, $4, $5, $6, $7::timestamptz, $8::timestamptz
) RETURNING jsonb_build_object('id', id);

-- name: LoginAdminFindSessionByAccess
-- params: access_token_hash text
SELECT jsonb_build_object(
    'session_id', s.id,
    'admin_id', a.id,
    'username', a.username,
    'email', a.email,
    'full_name', a.full_name,
    'role', a.role,
    'is_active', a.is_active,
    'access_expires_at', s.access_expires_at
) FROM admin_sessions s
JOIN admins a ON a.id = s.admin_id
WHERE s.access_token_hash = $1
  AND s.revoked_at IS NULL
  AND s.access_expires_at > NOW()
  AND a.is_active = TRUE;

-- name: LoginAdminFindSessionByRefresh
-- params: refresh_token_hash text
SELECT jsonb_build_object(
    'session_id', s.id,
    'admin_id', a.id,
    'device_id', s.device_id,
    'username', a.username,
    'email', a.email,
    'role', a.role,
    'is_active', a.is_active,
    'refresh_expires_at', s.refresh_expires_at
) FROM admin_sessions s
JOIN admins a ON a.id = s.admin_id
WHERE s.refresh_token_hash = $1
  AND s.revoked_at IS NULL
  AND s.refresh_expires_at > NOW()
  AND a.is_active = TRUE;

-- name: LoginAdminRevokeSession
-- params: session_id bigint
UPDATE admin_sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL;

-- name: LoginAdminRevokeAllForAdmin
-- params: admin_id bigint
UPDATE admin_sessions SET revoked_at = NOW() WHERE admin_id = $1 AND revoked_at IS NULL;

-- name: AdminProfile
-- params: admin_id bigint
SELECT jsonb_build_object(
    'id', id,
    'username', username,
    'email', email,
    'full_name', full_name,
    'role', role,
    'is_active', is_active,
    'last_login_at', last_login_at,
    'created_at', created_at
) FROM admins WHERE id = $1;

-- name: BootstrapAdminCreate
-- params: username text, email text, password_hash text, full_name text, role text
INSERT INTO admins (username, email, password_hash, full_name, role, is_active)
VALUES ($1, $2, $3, $4, $5, TRUE)
ON CONFLICT (email) DO UPDATE SET password_hash = excluded.password_hash, updated_at = NOW()
RETURNING jsonb_build_object('id', id, 'email', email);
