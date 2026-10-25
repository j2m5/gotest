BEGIN;

LOCK TABLE users IN ACCESS EXCLUSIVE MODE;
-- Explicit casts to varchar can silently truncate. Refuse rollback instead.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE char_length(email) > 32) THEN
        RAISE EXCEPTION 'auth hardening rollback: emails longer than 32 characters exist; shorten them explicitly before rolling back';
    END IF;
END $$;

-- SHA-256 is irreversible. Rollback revokes sessions and verification links;
-- account rows and credentials remain intact.
DELETE FROM sessions;
DELETE FROM email_verifications;
DROP INDEX email_verifications_user_id_idx;
DROP INDEX users_email_normalized_key;
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(32);
ALTER TABLE sessions DROP COLUMN expires_at;
ALTER TABLE email_verifications ALTER COLUMN expired_at DROP DEFAULT;
ALTER TABLE email_verifications ALTER COLUMN expired_at TYPE TIMESTAMP
    USING expired_at AT TIME ZONE current_setting('TimeZone');
ALTER TABLE email_verifications ALTER COLUMN expired_at SET DEFAULT (NOW() + INTERVAL '1 day');

COMMIT;
