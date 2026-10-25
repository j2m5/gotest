BEGIN;

-- Lock before checking duplicates so concurrent writes cannot invalidate the check.
LOCK TABLE users IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users GROUP BY lower(btrim(email)) HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'auth hardening: duplicate normalized emails exist; resolve them before migrating';
    END IF;
END $$;

ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(254);
CREATE UNIQUE INDEX users_email_normalized_key ON users (lower(btrim(email)));

ALTER TABLE sessions ADD COLUMN expires_at TIMESTAMPTZ;
-- Interpret legacy timestamps in the migration session's TimeZone.
UPDATE sessions SET expires_at = (created_at AT TIME ZONE current_setting('TimeZone')) + INTERVAL '168 hours';
ALTER TABLE sessions ALTER COLUMN expires_at SET NOT NULL;

ALTER TABLE email_verifications ALTER COLUMN expired_at DROP DEFAULT;
ALTER TABLE email_verifications ALTER COLUMN expired_at TYPE TIMESTAMPTZ
    USING expired_at AT TIME ZONE current_setting('TimeZone');
ALTER TABLE email_verifications ALTER COLUMN expired_at SET DEFAULT (NOW() + INTERVAL '24 hours');

-- PostgreSQL 12+ provides sha256(bytea) without requiring pgcrypto.
UPDATE sessions SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex');
UPDATE email_verifications SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex');

CREATE INDEX email_verifications_user_id_idx ON email_verifications(user_id);
COMMIT;
