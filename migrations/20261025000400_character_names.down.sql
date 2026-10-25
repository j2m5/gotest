BEGIN;
LOCK TABLE characters IN ACCESS EXCLUSIVE MODE;
ALTER TABLE characters ADD CONSTRAINT characters_name_key UNIQUE(name);
DROP INDEX characters_name_normalized_key;
DROP FUNCTION character_name_key(text);
DROP COLLATION character_name_unicode;
COMMIT;
