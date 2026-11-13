BEGIN;

ALTER TABLE battle_executions
 ADD COLUMN retry_count SMALLINT NOT NULL DEFAULT 0,
 ADD COLUMN retry_at TIMESTAMPTZ,
 ADD CONSTRAINT battle_executions_retry_check CHECK (
  retry_count BETWEEN 0 AND 6
  AND (retry_count=0)=(retry_at IS NULL)
  AND (retry_at IS NULL OR (isfinite(retry_at) AND next_action_at IS NOT NULL))
 );

-- PostgreSQL GREATEST ignores NULL: a normal encounter keeps its original due time.
DROP INDEX battle_executions_due_idx;
CREATE INDEX battle_executions_due_idx ON battle_executions
 (GREATEST(LEAST(next_action_at,expires_at),retry_at),battle_id)
 WHERE next_action_at IS NOT NULL;

COMMIT;
