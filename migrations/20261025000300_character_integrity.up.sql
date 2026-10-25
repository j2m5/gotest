BEGIN;

LOCK TABLE users, characters, classes IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM characters ch JOIN classes c ON c.id=ch.class_id
        WHERE ch.faction_id<>c.faction_id
    ) THEN
        RAISE EXCEPTION 'character integrity: class/faction mismatches exist; resolve them before migrating';
    END IF;
    IF EXISTS (
        SELECT 1 FROM users u LEFT JOIN characters ch ON ch.id=u.character_id AND ch.user_id=u.id
        WHERE u.character_id IS NOT NULL AND ch.id IS NULL
    ) THEN
        RAISE EXCEPTION 'character integrity: active characters must exist and belong to their accounts; resolve invalid selections before migrating';
    END IF;
END $$;

-- All validated selections refer to a SERIAL/INTEGER character ID, so the cast is lossless.
ALTER TABLE users ALTER COLUMN character_id TYPE INTEGER USING character_id::integer;
ALTER TABLE classes ADD CONSTRAINT classes_id_faction_id_key UNIQUE(id,faction_id);
ALTER TABLE characters
    ADD CONSTRAINT characters_user_id_id_key UNIQUE(user_id,id),
    ADD CONSTRAINT characters_class_faction_fkey FOREIGN KEY(class_id,faction_id)
        REFERENCES classes(id,faction_id),
    DROP CONSTRAINT characters_class_id_fkey;

-- MATCH SIMPLE allows an unselected (NULL) character. Account deletion still cascades
-- to its characters; deleting an active character requires clearing selection first.
ALTER TABLE users ADD CONSTRAINT users_active_character_owner_fkey
    FOREIGN KEY(id,character_id) REFERENCES characters(user_id,id);
COMMENT ON COLUMN users.character_id IS 'Active character shared by all account sessions; NULL when no character is selected';

COMMIT;
