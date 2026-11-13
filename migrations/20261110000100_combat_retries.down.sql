BEGIN;

DROP INDEX battle_executions_due_idx;
ALTER TABLE battle_executions DROP CONSTRAINT battle_executions_retry_check,
 DROP COLUMN retry_at, DROP COLUMN retry_count;
CREATE INDEX battle_executions_due_idx ON battle_executions
 (LEAST(next_action_at,expires_at),battle_id) WHERE next_action_at IS NOT NULL;

COMMIT;
