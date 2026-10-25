BEGIN;

LOCK TABLE users, sessions, email_verifications, factions, classes, abilities, characters IN ACCESS EXCLUSIVE MODE;

-- Legacy NOW()/clock_timestamp() values use the migration session's TimeZone.
-- The verification repository instead wrote one UTC timestamp to both user
-- columns. Preserve that known case before converting either column's type.
-- Run with the original metadata TimeZone; audit imported/manual dates first.
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING
        CASE WHEN updated_at = email_verified_at THEN updated_at AT TIME ZONE 'UTC'
             ELSE updated_at AT TIME ZONE current_setting('TimeZone') END,
    ALTER COLUMN email_verified_at TYPE TIMESTAMPTZ USING email_verified_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE sessions
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE email_verifications
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE factions
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE classes
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE abilities
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE characters
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN activated_at TYPE TIMESTAMPTZ USING activated_at AT TIME ZONE current_setting('TimeZone'),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

-- Metadata belongs to the database, including direct SQL and FK actions.
-- Use wall-clock time after any lock wait, not transaction-start NOW().
CREATE FUNCTION set_row_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at := clock_timestamp();
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER sessions_set_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER email_verifications_set_updated_at BEFORE UPDATE ON email_verifications
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER factions_set_updated_at BEFORE UPDATE ON factions
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER classes_set_updated_at BEFORE UPDATE ON classes
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER abilities_set_updated_at BEFORE UPDATE ON abilities
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER characters_set_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

COMMIT;
