BEGIN;

DROP INDEX sessions_presence_seen_idx;
ALTER TABLE sessions
    DROP CONSTRAINT sessions_last_seen_finite_check,
    DROP COLUMN last_seen_at,
    DROP COLUMN presence_character_id;

COMMIT;
