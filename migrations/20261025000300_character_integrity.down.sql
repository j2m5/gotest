BEGIN;
ALTER TABLE users DROP CONSTRAINT users_active_character_owner_fkey;
ALTER TABLE characters
    DROP CONSTRAINT characters_class_faction_fkey,
    DROP CONSTRAINT characters_user_id_id_key,
    ADD CONSTRAINT characters_class_id_fkey FOREIGN KEY(class_id) REFERENCES classes(id);
ALTER TABLE classes DROP CONSTRAINT classes_id_faction_id_key;
ALTER TABLE users ALTER COLUMN character_id TYPE BIGINT;
COMMENT ON COLUMN users.character_id IS NULL;
COMMIT;
