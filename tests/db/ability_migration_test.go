package db_test

import (
	"testing"
)

const abilityMigration = "20261108000200_ability_cooldowns"

func TestAbilityMigrationPreservesCustomBalanceAndResourceState(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, abilityMigration)
	mustExec(t, ctx, s, `UPDATE abilities SET damage=23,mp_cost=7,description='Custom effect' WHERE code='sith_warrior_basic_attack';
	 UPDATE abilities SET description='Наносит %d% урона оружием выбранной цели.' WHERE code='trooper_basic_attack';
	 INSERT INTO character_abilities(character_id,ability_id) SELECT 1,id FROM abilities WHERE code='sith_warrior_basic_attack'`)
	snapshot := func() string {
		var state string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_build_array(
		 (SELECT jsonb_agg(to_jsonb(c)-'global_cooldown_ready_at' ORDER BY id) FROM characters c),
		 (SELECT jsonb_agg(to_jsonb(a)-'cooldown_ms' ORDER BY id) FROM abilities a WHERE code='sith_warrior_basic_attack'),
		 (SELECT jsonb_agg(to_jsonb(ca)-'ready_at' ORDER BY character_id,ability_id) FROM character_abilities ca))::text`).Scan(&state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	before := snapshot()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, abilityMigration+"."+direction+".sql")
		if snapshot() != before {
			t.Fatalf("%s changed existing resources, ownership or custom balance", direction)
		}
	}
	var cooldown int
	var description string
	if err := s.pool.QueryRow(ctx, `SELECT cooldown_ms,description FROM abilities WHERE code='trooper_basic_attack'`).Scan(&cooldown, &description); err != nil || cooldown != 1500 || description != "Наносит базовый урон выбранной цели." {
		t.Fatalf("starter ability: %d %q %v", cooldown, description, err)
	}
	for _, tc := range []struct{ sql, constraint string }{
		{`UPDATE abilities SET cooldown_ms=-1`, "abilities_cooldown_ms_check"},
		{`UPDATE abilities SET cooldown_ms=86400001`, "abilities_cooldown_ms_check"},
		{`UPDATE abilities SET mp_cost=-1`, "abilities_effect_amounts_check"},
		{`UPDATE abilities SET damage=-1`, "abilities_effect_amounts_check"},
		{`UPDATE abilities SET heal=-1`, "abilities_effect_amounts_check"},
		{`UPDATE characters SET global_cooldown_ready_at='infinity'`, "characters_global_cooldown_ready_at_check"},
		{`UPDATE character_abilities SET ready_at='-infinity'`, "character_abilities_ready_at_check"},
	} {
		_, err := s.pool.Exec(ctx, tc.sql)
		requireConstraint(t, err, "23514", tc.constraint)
	}
}
