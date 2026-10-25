BEGIN;

LOCK TABLE characters IN ACCESS EXCLUSIVE MODE;

-- Pin case conversion to Unicode even when the database uses LC_CTYPE=C.
CREATE COLLATION character_name_unicode (provider=icu, locale='und', deterministic=true);

-- Keep the key aligned with Go's NFC + strings.TrimSpace + ASCII-space collapse.
-- The explicit trim set is Unicode White_Space, which strings.TrimSpace uses.
CREATE FUNCTION character_name_key(text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
RETURN lower(normalize(regexp_replace(btrim($1,
    U&'\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000\2001\2002\2003\2004\2005\2006\2007\2008\2009\200A\2028\2029\202F\205F\3000'),
    ' +', ' ', 'g'), NFC) COLLATE character_name_unicode);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM characters GROUP BY character_name_key(name) HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'character names: duplicate normalized names exist; resolve them before migrating';
    END IF;
END $$;

CREATE UNIQUE INDEX characters_name_normalized_key ON characters (character_name_key(name));
ALTER TABLE characters DROP CONSTRAINT characters_name_key;

COMMIT;
