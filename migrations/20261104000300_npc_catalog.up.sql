BEGIN;

CREATE TABLE npc_templates (
    id SERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    name VARCHAR(100) NOT NULL CHECK (char_length(btrim(name)) > 0),
    description TEXT NOT NULL DEFAULT '',
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('humanoid','droid','beast')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER npc_templates_updated_at BEFORE UPDATE ON npc_templates
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

CREATE TABLE location_npcs (
    id SERIAL PRIMARY KEY,
    location_id INT NOT NULL REFERENCES locations(id),
    npc_template_id INT NOT NULL REFERENCES npc_templates(id),
    code VARCHAR(64) NOT NULL CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    position SMALLINT NOT NULL CHECK (position > 0),
    level SMALLINT NOT NULL REFERENCES levels(level),
    rank VARCHAR(16) NOT NULL CHECK (rank IN ('normal','strong','elite')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (location_id,code),
    UNIQUE (location_id,position)
);
CREATE INDEX location_npcs_template_idx ON location_npcs(npc_template_id);
CREATE TRIGGER location_npcs_updated_at BEFORE UPDATE ON location_npcs
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

-- Persistent shared identities. Battle ownership and respawn fields belong to
-- the battle migration, where ownership can reference an actual battle row.
CREATE TABLE npc_spawns (
    id SERIAL PRIMARY KEY,
    location_npc_id INT NOT NULL REFERENCES location_npcs(id),
    slot SMALLINT NOT NULL CHECK (slot > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (location_npc_id,slot)
);
CREATE TRIGGER npc_spawns_updated_at BEFORE UPDATE ON npc_spawns
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

COMMIT;
