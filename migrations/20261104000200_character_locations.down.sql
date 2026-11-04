BEGIN;
DROP TRIGGER characters_start_location ON characters;
DROP FUNCTION set_character_start_location();
ALTER TABLE characters DROP COLUMN location_id;
DROP TABLE locations;
COMMIT;
