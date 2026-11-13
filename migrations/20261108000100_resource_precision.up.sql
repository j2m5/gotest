BEGIN;

ALTER TABLE characters
    ADD COLUMN hp_regen_remainder INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN mp_regen_remainder INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN resources_revision BIGINT NOT NULL DEFAULT 0 CHECK (resources_revision >= 0),
    ADD CONSTRAINT characters_hp_remainder_check CHECK (hp_regen_remainder >= 0 AND hp_regen_remainder < 300000 AND (hp_current < hp OR hp_regen_remainder = 0)),
    ADD CONSTRAINT characters_mp_remainder_check CHECK (mp_regen_remainder >= 0 AND mp_regen_remainder < 60000 AND (mp_current < mp OR mp_regen_remainder = 0));

CREATE OR REPLACE FUNCTION track_character_resource_changes() RETURNS trigger AS $$
DECLARE
    changed_at TIMESTAMPTZ := clock_timestamp();
BEGIN
    -- The transactional Go writer supplies a complete resource snapshot and
    -- increments this revision, including two writes in the same millisecond.
    IF NEW.resources_revision IS DISTINCT FROM OLD.resources_revision THEN
        IF NEW.resources_revision <> OLD.resources_revision + 1 THEN
            RAISE EXCEPTION 'resource revision must advance by one';
        END IF;
        RETURN NEW;
    END IF;
    -- Preserve the historical behavior of direct baseline/capacity replacement.
    IF NEW.hp_current IS DISTINCT FROM OLD.hp_current OR NEW.hp IS DISTINCT FROM OLD.hp THEN
        NEW.hp_updated_at := changed_at;
        NEW.hp_regen_remainder := 0;
    END IF;
    IF NEW.mp_current IS DISTINCT FROM OLD.mp_current OR NEW.mp IS DISTINCT FROM OLD.mp THEN
        NEW.mp_updated_at := changed_at;
        NEW.mp_regen_remainder := 0;
    END IF;
    IF NEW.hp_current IS DISTINCT FROM OLD.hp_current OR NEW.hp IS DISTINCT FROM OLD.hp
       OR NEW.mp_current IS DISTINCT FROM OLD.mp_current OR NEW.mp IS DISTINCT FROM OLD.mp THEN
        NEW.resources_revision := OLD.resources_revision + 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMIT;
