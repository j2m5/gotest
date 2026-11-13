BEGIN;

ALTER TABLE abilities
    ADD COLUMN cooldown_ms INTEGER NOT NULL DEFAULT 1500 CHECK (cooldown_ms BETWEEN 0 AND 86400000),
    ADD CONSTRAINT abilities_effect_amounts_check CHECK (damage >= 0 AND heal >= 0 AND mp_cost >= 0);
ALTER TABLE character_abilities
    ADD COLUMN ready_at TIMESTAMPTZ CHECK (ready_at IS NULL OR isfinite(ready_at));
ALTER TABLE characters
    ADD COLUMN global_cooldown_ready_at TIMESTAMPTZ CHECK (global_cooldown_ready_at IS NULL OR isfinite(global_cooldown_ready_at));

-- Correct only the shipped placeholder; preserve custom descriptions/balance.
UPDATE abilities SET description='Наносит базовый урон выбранной цели.'
WHERE description='Наносит %d% урона оружием выбранной цели.' AND code IN (
    'sith_warrior_basic_attack','sith_inquisitor_basic_attack','imperial_agent_basic_attack','bounty_hunter_basic_attack',
    'jedi_knight_basic_attack','jedi_consular_basic_attack','smuggler_basic_attack','trooper_basic_attack');

COMMIT;
