BEGIN;

LOCK TABLE users, sessions, email_verifications, factions, classes, abilities, characters IN ACCESS EXCLUSIVE MODE;

DROP TRIGGER users_set_updated_at ON users;
DROP TRIGGER sessions_set_updated_at ON sessions;
DROP TRIGGER email_verifications_set_updated_at ON email_verifications;
DROP TRIGGER factions_set_updated_at ON factions;
DROP TRIGGER classes_set_updated_at ON classes;
DROP TRIGGER abilities_set_updated_at ON abilities;
DROP TRIGGER characters_set_updated_at ON characters;
DROP FUNCTION set_row_updated_at();

-- Restore the previous conventions using the same TimeZone as the up migration.
-- Expiry columns were already timestamptz and deliberately remain unchanged.
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING
        CASE WHEN updated_at = email_verified_at THEN updated_at AT TIME ZONE 'UTC'
             ELSE updated_at AT TIME ZONE current_setting('TimeZone') END,
    ALTER COLUMN email_verified_at TYPE TIMESTAMP USING email_verified_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE sessions
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE email_verifications
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE factions
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE classes
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE abilities
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE characters
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN activated_at TYPE TIMESTAMP USING activated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

COMMIT;
