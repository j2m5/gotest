BEGIN;

-- Cleanup acquires a battle row lock before touching archive/event data. Wait
-- for those transactions and exclude new ones before checking irreversible work.
LOCK TABLE battles IN EXCLUSIVE MODE;

DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM battle_log_archives) THEN
  RAISE EXCEPTION 'cannot restore cleared combat logs; restore a backup instead';
 END IF;
END $$;

DROP TRIGGER battle_events_delete ON battle_events;
DROP FUNCTION guard_battle_event_delete();
DROP TRIGGER battle_commands_insert ON battle_commands;
DROP FUNCTION guard_battle_command_insert();
ALTER TABLE battle_commands DROP CONSTRAINT battle_commands_sequence_check;
ALTER TABLE battle_commands ADD CONSTRAINT battle_commands_battle_id_event_sequence_fkey
 FOREIGN KEY(battle_id,event_sequence) REFERENCES battle_events(battle_id,sequence);
DROP TABLE battle_log_archives;
DROP FUNCTION guard_battle_log_archive_insert();
DROP INDEX battles_ended_idx;

COMMIT;
