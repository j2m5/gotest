BEGIN;

ALTER TABLE battles DROP CONSTRAINT battles_status_check;
ALTER TABLE battles ADD CONSTRAINT battles_status_check CHECK (status IN ('active','retreated','defeated','won','expired'));
ALTER TABLE battles DROP CONSTRAINT battles_released_health_check;
ALTER TABLE battles ADD CONSTRAINT battles_released_health_check CHECK (
 status='active' OR (status='won' AND npc_hp_current=0) OR (status IN ('retreated','defeated','expired') AND npc_hp_current=npc_hp));

CREATE TABLE battle_executions (
 battle_id BIGINT PRIMARY KEY REFERENCES battles(id),
 rules JSONB NOT NULL CHECK (jsonb_typeof(rules)='object'),
 pattern_step INTEGER NOT NULL DEFAULT 0 CHECK (pattern_step>=0),
 next_action_at TIMESTAMPTZ CHECK (next_action_at IS NULL OR isfinite(next_action_at)),
 expires_at TIMESTAMPTZ NOT NULL CHECK (isfinite(expires_at)),
 event_sequence INTEGER NOT NULL DEFAULT 0 CHECK (event_sequence>=0),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(created_at)),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(updated_at)),
 CHECK (expires_at>created_at AND expires_at<=created_at+INTERVAL '10 minutes'),
 CHECK (next_action_at IS NULL OR next_action_at>created_at)
);
CREATE INDEX battle_executions_due_idx ON battle_executions(LEAST(next_action_at,expires_at),battle_id) WHERE next_action_at IS NOT NULL;

CREATE TABLE battle_events (
 battle_id BIGINT NOT NULL REFERENCES battles(id),
 sequence INTEGER NOT NULL CHECK (sequence>0),
 occurred_at TIMESTAMPTZ NOT NULL CHECK (isfinite(occurred_at)),
 payload JSONB NOT NULL CHECK (jsonb_typeof(payload)='object'),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(created_at)),
 PRIMARY KEY(battle_id,sequence)
);

CREATE TABLE battle_commands (
 battle_id BIGINT NOT NULL REFERENCES battles(id),
 command_key VARCHAR(128) NOT NULL CHECK (command_key ~ '^[a-zA-Z0-9]' AND command_key !~ '[^a-zA-Z0-9:._/-]'),
 ability_id INTEGER NOT NULL REFERENCES abilities(id),
 event_sequence INTEGER NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(created_at)),
 PRIMARY KEY(battle_id,command_key),
 FOREIGN KEY(battle_id,event_sequence) REFERENCES battle_events(battle_id,sequence)
);
CREATE INDEX battle_commands_ability_idx ON battle_commands(ability_id);

CREATE FUNCTION guard_battle_execution_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF ROW(NEW.battle_id,NEW.rules,NEW.expires_at,NEW.created_at)
    IS DISTINCT FROM ROW(OLD.battle_id,OLD.rules,OLD.expires_at,OLD.created_at)
    OR NEW.pattern_step<OLD.pattern_step OR NEW.event_sequence<OLD.event_sequence
    OR (OLD.next_action_at IS NULL AND NEW IS DISTINCT FROM OLD) THEN
  RAISE EXCEPTION 'combat snapshot and finished execution are immutable' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER battle_executions_guard BEFORE UPDATE ON battle_executions FOR EACH ROW EXECUTE FUNCTION guard_battle_execution_update();
CREATE TRIGGER battle_executions_updated_at BEFORE UPDATE ON battle_executions FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE FUNCTION guard_combat_record_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'combat records are immutable' USING ERRCODE='23514';
END;
$$;
CREATE TRIGGER battle_events_guard BEFORE UPDATE ON battle_events FOR EACH ROW EXECUTE FUNCTION guard_combat_record_update();
CREATE TRIGGER battle_commands_guard BEFORE UPDATE ON battle_commands FOR EACH ROW EXECUTE FUNCTION guard_combat_record_update();

COMMIT;
