BEGIN;

CREATE OR REPLACE FUNCTION track_character_resource_changes() RETURNS trigger AS $$
DECLARE
    changed_at TIMESTAMPTZ := clock_timestamp();
BEGIN
    IF NEW.hp_current IS DISTINCT FROM OLD.hp_current OR NEW.hp IS DISTINCT FROM OLD.hp THEN
        NEW.hp_updated_at := changed_at;
    END IF;
    IF NEW.mp_current IS DISTINCT FROM OLD.mp_current OR NEW.mp IS DISTINCT FROM OLD.mp THEN
        NEW.mp_updated_at := changed_at;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE characters
    DROP COLUMN hp_regen_remainder,
    DROP COLUMN mp_regen_remainder,
    DROP COLUMN resources_revision;

COMMIT;
