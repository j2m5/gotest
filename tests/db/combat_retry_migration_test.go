package db_test

import "testing"

const combatRetryMigration = "20261110000100_combat_retries"

func TestCombatRetryMigrationPreservesExistingEncounter(t *testing.T) {
	f := newExecutionFixture(t)
	if _, err := f.cast("before-migration"); err != nil {
		t.Fatal(err)
	}
	before := combatEffectsSnapshot(t, f)
	var updated string
	if err := f.s.pool.QueryRow(f.ctx, `SELECT updated_at::text FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&updated); err != nil {
		t.Fatal(err)
	}
	for _, direction := range []string{"down", "up", "down", "up"} {
		executeMigration(t, f.ctx, f.s, combatRetryMigration+"."+direction+".sql")
		var current string
		if err := f.s.pool.QueryRow(f.ctx, `SELECT updated_at::text FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&current); err != nil || current != updated || combatEffectsSnapshot(t, f) != before {
			t.Fatal("migration changed game data or timestamps", direction, current, updated, err)
		}
	}
}

func TestCombatRetryMigrationGuards(t *testing.T) {
	f := newExecutionFixture(t)
	for _, assignment := range []string{
		"retry_count=-1", "retry_count=7", "retry_count=1", "retry_at=clock_timestamp()",
		"retry_count=1,retry_at='infinity'", "retry_count=1,retry_at='-infinity'",
		"retry_count=1,retry_at=clock_timestamp(),next_action_at=NULL",
	} {
		_, err := f.s.pool.Exec(f.ctx, "UPDATE battle_executions SET "+assignment+" WHERE battle_id=$1", f.battle.ID)
		requireConstraint(t, err, "23514", "battle_executions_retry_check")
	}
	if _, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID); err != nil {
		t.Fatal(err)
	}
	_, err := f.s.pool.Exec(f.ctx, `UPDATE battle_executions SET retry_count=1,retry_at=clock_timestamp() WHERE battle_id=$1`, f.battle.ID)
	requireConstraint(t, err, "23514", "")
}
