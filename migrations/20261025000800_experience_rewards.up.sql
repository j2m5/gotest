BEGIN;

-- One immutable outcome per character and source event. History uses plain
-- level numbers so later catalog balancing cannot invalidate old outcomes.
CREATE TABLE character_experience_rewards (
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    reward_key VARCHAR(128) COLLATE "C" NOT NULL,
    requested_amount BIGINT NOT NULL,
    applied_amount INTEGER NOT NULL,
    experience_before INTEGER NOT NULL,
    experience_after INTEGER NOT NULL,
    level_before SMALLINT NOT NULL,
    level_after SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (character_id, reward_key),
    CONSTRAINT character_experience_rewards_key_check
        CHECK (reward_key ~ '^[A-Za-z0-9][A-Za-z0-9:._/-]{0,127}$'),
    CONSTRAINT character_experience_rewards_requested_check
        CHECK (requested_amount > 0),
    CONSTRAINT character_experience_rewards_applied_check
        CHECK (applied_amount >= 0 AND applied_amount <= requested_amount),
    CONSTRAINT character_experience_rewards_experience_check
        CHECK (experience_before >= 0 AND experience_after >= experience_before
            AND experience_after::BIGINT - experience_before::BIGINT = applied_amount),
    CONSTRAINT character_experience_rewards_levels_check
        CHECK (level_before > 0 AND level_after >= level_before)
);

COMMIT;
