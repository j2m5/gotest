BEGIN;

-- Discarding recorded events would allow their rewards to be applied again.
LOCK TABLE character_experience_rewards IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM character_experience_rewards) THEN
        RAISE EXCEPTION 'experience rewards: cannot roll back a nonempty reward ledger';
    END IF;
END $$;

DROP TABLE character_experience_rewards;

COMMIT;
