-- 001_admins_and_sessions.up.sql
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'ADMIN',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_admins_username UNIQUE (username),
    CONSTRAINT uq_admins_email UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS admin_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    access_token_hash TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    access_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_admin_sessions_access UNIQUE (access_token_hash),
    CONSTRAINT uq_admin_sessions_refresh UNIQUE (refresh_token_hash)
);

CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin_id ON admin_sessions (admin_id) WHERE revoked_at IS NULL;
