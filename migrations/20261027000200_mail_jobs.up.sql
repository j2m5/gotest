CREATE TABLE mail_jobs (
    id text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{64}$'),
    kind text NOT NULL CHECK (kind IN ('verification', 'password_reset', 'password_changed', 'reset_request')),
    key_id text NOT NULL CHECK (key_id ~ '^[0-9a-f]{64}$'),
    payload bytea NOT NULL CHECK (octet_length(payload) > 0),
    user_id integer REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    expires_at timestamptz NOT NULL,
    available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 8),
    lease_token text,
    leased_until timestamptz,
    failed_at timestamptz,
    CHECK ((kind = 'reset_request') = (user_id IS NULL)),
    CHECK ((lease_token IS NULL) = (leased_until IS NULL)),
    CHECK (lease_token IS NULL OR lease_token ~ '^[0-9a-f]{64}$'),
    CHECK (failed_at IS NULL OR leased_until IS NULL)
);

CREATE INDEX mail_jobs_due_idx ON mail_jobs (available_at, created_at, id)
    WHERE failed_at IS NULL AND attempts < 8;
CREATE INDEX mail_jobs_expiry_idx ON mail_jobs (expires_at, id);
CREATE INDEX mail_jobs_user_kind_idx ON mail_jobs (user_id, kind);
