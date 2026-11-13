BEGIN;

CREATE TABLE battle_log_archives (
 battle_id BIGINT PRIMARY KEY REFERENCES battles(id),
 final_sequence INTEGER NOT NULL CHECK (final_sequence>0),
 cleared_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp() CHECK (isfinite(cleared_at)),
 FOREIGN KEY(battle_id,final_sequence) REFERENCES battle_events(battle_id,sequence)
);
CREATE INDEX battles_ended_idx ON battles(ended_at,id) WHERE status<>'active';

CREATE FUNCTION guard_battle_log_archive_insert() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE battle_status text; ended timestamptz; final integer; next_action timestamptz;
BEGIN
 SELECT b.status,b.ended_at,e.event_sequence,e.next_action_at
 INTO battle_status,ended,final,next_action
 FROM battles b JOIN battle_executions e ON e.battle_id=b.id
 WHERE b.id=NEW.battle_id FOR UPDATE OF b;
 IF NOT FOUND OR battle_status='active' OR ended IS NULL OR next_action IS NOT NULL
    OR final<>NEW.final_sequence OR NEW.cleared_at<ended
    OR NOT EXISTS(SELECT 1 FROM battle_events v WHERE v.battle_id=NEW.battle_id
       AND v.sequence=final AND v.payload->>'kind'=battle_status AND v.occurred_at=ended) THEN
  RAISE EXCEPTION 'only a completed combat log with its final event can be cleared' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER battle_log_archives_insert BEFORE INSERT ON battle_log_archives
 FOR EACH ROW EXECUTE FUNCTION guard_battle_log_archive_insert();
CREATE TRIGGER battle_log_archives_guard BEFORE UPDATE OR DELETE ON battle_log_archives
 FOR EACH ROW EXECUTE FUNCTION guard_combat_record_update();

-- Receipts survive the details they acknowledge. A new receipt must still be
-- created against an existing event; the battle lock serializes it with cleanup.
ALTER TABLE battle_commands DROP CONSTRAINT battle_commands_battle_id_event_sequence_fkey;
ALTER TABLE battle_commands ADD CONSTRAINT battle_commands_sequence_check CHECK (event_sequence>0);
CREATE FUNCTION guard_battle_command_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 PERFORM 1 FROM battles WHERE id=NEW.battle_id FOR KEY SHARE;
 IF EXISTS(SELECT 1 FROM battle_log_archives WHERE battle_id=NEW.battle_id) THEN
  RAISE EXCEPTION 'cleared combat logs cannot accept new commands' USING ERRCODE='23514';
 END IF;
 PERFORM 1 FROM battle_events WHERE battle_id=NEW.battle_id AND sequence=NEW.event_sequence FOR KEY SHARE;
 IF NOT FOUND THEN
  RAISE EXCEPTION 'combat command event does not exist' USING ERRCODE='23503';
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER battle_commands_insert BEFORE INSERT ON battle_commands
 FOR EACH ROW EXECUTE FUNCTION guard_battle_command_insert();

CREATE FUNCTION guard_battle_event_delete() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NOT EXISTS(SELECT 1 FROM battle_log_archives WHERE battle_id=OLD.battle_id) THEN
  RAISE EXCEPTION 'combat details require a retained outcome before deletion' USING ERRCODE='23514';
 END IF;
 RETURN OLD;
END;
$$;
CREATE TRIGGER battle_events_delete BEFORE DELETE ON battle_events
 FOR EACH ROW EXECUTE FUNCTION guard_battle_event_delete();

COMMIT;
