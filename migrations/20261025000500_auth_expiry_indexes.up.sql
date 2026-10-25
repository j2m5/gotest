BEGIN;

CREATE INDEX sessions_expires_at_id_idx ON sessions (expires_at, id);
CREATE INDEX email_verifications_expired_at_id_idx ON email_verifications (expired_at, id);

COMMIT;
