BEGIN;
DROP TRIGGER characters_resource_changes ON characters;
DROP FUNCTION track_character_resource_changes();
ALTER TABLE characters
    DROP CONSTRAINT characters_resource_times_finite_check,
    DROP COLUMN hp_updated_at,
    DROP COLUMN mp_updated_at;
COMMIT;
