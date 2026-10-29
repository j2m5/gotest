BEGIN;

-- Existing amounts remain untouched. Recovery starts when this migration is
-- applied, rather than granting time since an unrelated character update.
ALTER TABLE characters
    ADD COLUMN hp_updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    ADD COLUMN mp_updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    ADD CONSTRAINT characters_resource_times_finite_check CHECK (isfinite(hp_updated_at) AND isfinite(mp_updated_at));

CREATE FUNCTION track_character_resource_changes() RETURNS trigger AS $$
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

CREATE TRIGGER characters_resource_changes
    BEFORE UPDATE OF hp_current, hp, mp_current, mp ON characters
    FOR EACH ROW EXECUTE FUNCTION track_character_resource_changes();

COMMIT;
