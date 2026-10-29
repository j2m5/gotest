BEGIN;

ALTER TABLE sessions
    ADD COLUMN presence_character_id INTEGER REFERENCES characters(id) ON DELETE SET NULL,
    ADD COLUMN last_seen_at TIMESTAMPTZ,
    ADD CONSTRAINT sessions_last_seen_finite_check CHECK (last_seen_at IS NULL OR isfinite(last_seen_at));

CREATE INDEX sessions_presence_seen_idx ON sessions(last_seen_at, presence_character_id)
    WHERE presence_character_id IS NOT NULL;

COMMIT;
