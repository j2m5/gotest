BEGIN;

DROP INDEX email_verifications_expired_at_id_idx;
DROP INDEX sessions_expires_at_id_idx;

COMMIT;
