BEGIN;

CREATE TABLE password_resets (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE CHECK (token ~ '^[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX password_resets_expires_at_id_idx ON password_resets (expires_at, id);

COMMIT;
