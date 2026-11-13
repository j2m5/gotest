BEGIN;

ALTER TABLE characters DROP COLUMN global_cooldown_ready_at;
ALTER TABLE character_abilities DROP COLUMN ready_at;
ALTER TABLE abilities DROP COLUMN cooldown_ms, DROP CONSTRAINT abilities_effect_amounts_check;

COMMIT;
