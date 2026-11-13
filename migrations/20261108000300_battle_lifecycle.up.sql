BEGIN;

ALTER TABLE npc_spawns ADD COLUMN respawn_at TIMESTAMPTZ
    CHECK (respawn_at IS NULL OR isfinite(respawn_at));

CREATE TABLE battles (
    id BIGSERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    npc_spawn_id INTEGER NOT NULL REFERENCES npc_spawns(id),
    location_id INTEGER NOT NULL REFERENCES locations(id),
    start_key VARCHAR(128) NOT NULL CHECK (start_key ~ '^[a-zA-Z0-9]' AND start_key !~ '[^a-zA-Z0-9:._/-]'),
    status VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active','retreated','defeated')),
    npc_template_id INTEGER NOT NULL REFERENCES npc_templates(id),
    npc_name VARCHAR(100) NOT NULL CHECK (char_length(btrim(npc_name)) > 0),
    npc_level SMALLINT NOT NULL CHECK (npc_level > 0),
    npc_rank VARCHAR(16) NOT NULL CHECK (npc_rank IN ('normal','strong','elite')),
    npc_hp INTEGER NOT NULL CHECK (npc_hp > 0),
    npc_hp_current INTEGER NOT NULL CONSTRAINT battles_npc_hp_current_check CHECK (npc_hp_current >= 0 AND npc_hp_current <= npc_hp),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(created_at)),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(updated_at)),
    ended_at TIMESTAMPTZ CONSTRAINT battles_ended_at_check CHECK (ended_at IS NULL OR (isfinite(ended_at) AND ended_at >= created_at)),
    UNIQUE (character_id,start_key),
    CONSTRAINT battles_status_time_check CHECK ((status='active') = (ended_at IS NULL)),
    CONSTRAINT battles_released_health_check CHECK (status='active' OR npc_hp_current=npc_hp)
);
CREATE UNIQUE INDEX battles_active_character_idx ON battles(character_id) WHERE status='active';
CREATE UNIQUE INDEX battles_active_spawn_idx ON battles(npc_spawn_id) WHERE status='active';
CREATE INDEX battles_spawn_idx ON battles(npc_spawn_id);
CREATE INDEX battles_location_idx ON battles(location_id);
CREATE INDEX battles_template_idx ON battles(npc_template_id);

CREATE FUNCTION guard_battle_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF ROW(NEW.id,NEW.character_id,NEW.npc_spawn_id,NEW.location_id,NEW.start_key,
           NEW.npc_template_id,NEW.npc_name,NEW.npc_level,NEW.npc_rank,NEW.npc_hp,NEW.created_at)
       IS DISTINCT FROM
       ROW(OLD.id,OLD.character_id,OLD.npc_spawn_id,OLD.location_id,OLD.start_key,
           OLD.npc_template_id,OLD.npc_name,OLD.npc_level,OLD.npc_rank,OLD.npc_hp,OLD.created_at) THEN
        RAISE EXCEPTION 'battle identity and initial parameters are immutable' USING ERRCODE='23514';
    END IF;
    IF OLD.status <> 'active' AND NEW IS DISTINCT FROM OLD THEN
        RAISE EXCEPTION 'finished battle is immutable' USING ERRCODE='23514';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER battles_guard_update BEFORE UPDATE ON battles FOR EACH ROW EXECUTE FUNCTION guard_battle_update();
CREATE TRIGGER battles_updated_at BEFORE UPDATE ON battles FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

COMMIT;
