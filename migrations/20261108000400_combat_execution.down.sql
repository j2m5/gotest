BEGIN;
LOCK TABLE battles,battle_executions,battle_events,battle_commands IN ACCESS EXCLUSIVE MODE;
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM battle_executions) OR EXISTS(SELECT 1 FROM battle_events)
    OR EXISTS(SELECT 1 FROM battle_commands) OR EXISTS(SELECT 1 FROM battles WHERE status IN ('won','expired')) THEN
  RAISE EXCEPTION 'cannot roll back nonempty combat execution';
 END IF;
END $$;
DROP TABLE battle_commands;
DROP TABLE battle_events;
DROP TABLE battle_executions;
DROP FUNCTION guard_combat_record_update();
DROP FUNCTION guard_battle_execution_update();
ALTER TABLE battles DROP CONSTRAINT battles_status_check;
ALTER TABLE battles ADD CONSTRAINT battles_status_check CHECK (status IN ('active','retreated','defeated'));
ALTER TABLE battles DROP CONSTRAINT battles_released_health_check;
ALTER TABLE battles ADD CONSTRAINT battles_released_health_check CHECK (status='active' OR npc_hp_current=npc_hp);
COMMIT;
