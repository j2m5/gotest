package db_test

import (
	"strings"
	"testing"
)

const executionMigration = "20261108000400_combat_execution"

func TestCombatExecutionMigrationRoundTripAndGuards(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, executionMigration)
	before := characterStatsSnapshot(t, ctx, s)
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, executionMigration+"."+direction+".sql")
		if characterStatsSnapshot(t, ctx, s) != before {
			t.Fatal("migration changed characters")
		}
	}
}

func TestCombatExecutionMigrationProtectsHistoryAndRules(t *testing.T) {
	f := newExecutionFixture(t)
	_, err := f.s.pool.Exec(f.ctx, `UPDATE battle_executions SET rules=rules||'{"hp":1}' WHERE battle_id=$1`, f.battle.ID)
	requireConstraint(t, err, "23514", "")
	_, err = f.s.pool.Exec(f.ctx, `UPDATE battle_events SET payload='{}' WHERE battle_id=$1`, f.battle.ID)
	requireConstraint(t, err, "23514", "")
	conn, err := f.s.pool.Acquire(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	_, err = conn.Exec(f.ctx, migrationSQL(t, executionMigration+".down.sql"))
	if err == nil || !strings.Contains(err.Error(), "cannot roll back nonempty combat execution") {
		t.Fatal(err)
	}
	if _, err = conn.Exec(f.ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if len(f.view(t).Events) != 1 {
		t.Fatal("history lost")
	}
}
