CREATE TABLE planets (
    id SERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    name VARCHAR(100) NOT NULL CHECK (char_length(btrim(name)) > 0),
    name_en VARCHAR(100) NOT NULL CHECK (char_length(btrim(name_en)) > 0),
    description TEXT,
    core_faction_id INT REFERENCES factions(id),
    controlling_faction_id INT REFERENCES factions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT planets_core_control CHECK (
        core_faction_id IS NULL OR
        (controlling_faction_id IS NOT NULL AND controlling_faction_id = core_faction_id)
    )
);

-- Core membership is part of the world definition, not capturable ownership.
-- Protect the rule even if a future capture accidentally updates both fields.
CREATE FUNCTION protect_planet_core() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.core_faction_id IS DISTINCT FROM OLD.core_faction_id THEN
        RAISE EXCEPTION 'planet core membership is immutable' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER planets_protect_core BEFORE UPDATE OF core_faction_id ON planets
    FOR EACH ROW EXECUTE FUNCTION protect_planet_core();
CREATE TRIGGER planets_updated_at BEFORE UPDATE ON planets
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

CREATE INDEX planets_controlling_faction_idx ON planets(controlling_faction_id);
