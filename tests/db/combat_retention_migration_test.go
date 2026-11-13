package db_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"gotest/internal/maintenance"
)

const retentionMigration = "20261113000100_battle_log_retention"

func TestCombatLogDowngradeWaitsForConcurrentPurge(t *testing.T) {
	f := newExecutionFixture(t)
	result, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID)
	if err != nil {
		t.Fatal(err)
	}
	f.battle = result.Battle
	f = cloneFinishedLog(t, f, 40) // No command receipts: restoring their FK cannot detect data loss.
	cleaner, err := f.s.pool.Begin(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cleaner.Rollback(context.Background())
	if _, err = cleaner.Exec(f.ctx, `SELECT id FROM battles WHERE id=$1 FOR UPDATE`, f.battle.ID); err != nil {
		t.Fatal(err)
	}
	// Hold the event-table lock modes before downgrade queues its DDL. This lets
	// this transaction finish the purge at a deterministic point after the old DO.
	if _, err = cleaner.Exec(f.ctx, `SELECT payload FROM battle_events WHERE battle_id=$1`, f.battle.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = cleaner.Exec(f.ctx, `SELECT sequence FROM battle_events WHERE battle_id=$1 FOR KEY SHARE`, f.battle.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = cleaner.Exec(f.ctx, `LOCK TABLE battle_events IN ROW EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}
	conn, err := f.s.pool.Acquire(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	pid := conn.Conn().PgConn().PID()
	ctx, cancel := context.WithTimeout(f.ctx, 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	downSQL := migrationSQL(t, retentionMigration+".down.sql")
	go func() {
		_, err := conn.Exec(ctx, downSQL)
		if err != nil {
			_, _ = conn.Exec(context.Background(), "ROLLBACK")
		}
		done <- err
		close(done)
	}()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var waiting bool
		if err := f.s.pool.QueryRow(f.ctx, `SELECT EXISTS(SELECT 1 FROM pg_locks WHERE pid=$1 AND NOT granted
 AND relation IN ('battles'::regclass,'battle_events'::regclass))`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("downgrade did not reach its lock")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := cleaner.Exec(f.ctx, `INSERT INTO battle_log_archives(battle_id,final_sequence) VALUES($1,2)`, f.battle.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := cleaner.Exec(f.ctx, `DELETE FROM battle_events WHERE battle_id=$1 AND sequence<>2`, f.battle.ID); err != nil {
		t.Fatal(err)
	}
	if err := cleaner.Commit(f.ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err == nil || !strings.Contains(err.Error(), "cannot restore cleared combat logs") {
		t.Fatal("downgrade erased a concurrently cleared journal", err)
	}
	logCounts(t, f, 1, 1)
}

func TestCombatLogRetentionMigrationRoundTrip(t *testing.T) {
	f := wonLogFixture(t)
	before := combatEffectsSnapshot(t, f)
	for _, direction := range []string{"down", "up", "down", "up"} {
		executeMigration(t, f.ctx, f.s, retentionMigration+"."+direction+".sql")
		if combatEffectsSnapshot(t, f) != before {
			t.Fatal("migration changed encounter or rewards", direction)
		}
	}
	old := cloneFinishedLog(t, f, 40)
	if _, err := maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true}); err != nil {
		t.Fatal(err)
	}
	conn, err := f.s.pool.Acquire(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	_, err = conn.Exec(f.ctx, migrationSQL(t, retentionMigration+".down.sql"))
	if err == nil || !strings.Contains(err.Error(), "cannot restore cleared combat logs") {
		t.Fatal("unsafe downgrade", err)
	}
	if _, err := conn.Exec(f.ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	logCounts(t, old, 1, 1)
}

func TestCombatLogRetentionDatabaseGuards(t *testing.T) {
	f := newExecutionFixture(t)
	_, err := f.s.pool.Exec(f.ctx, `INSERT INTO battle_log_archives(battle_id,final_sequence) VALUES($1,1)`, f.battle.ID)
	requireConstraint(t, err, "23514", "")
	_, err = f.s.pool.Exec(f.ctx, `DELETE FROM battle_events WHERE battle_id=$1`, f.battle.ID)
	requireConstraint(t, err, "23514", "")
	_, err = f.s.pool.Exec(f.ctx, `INSERT INTO battle_commands(battle_id,command_key,ability_id,event_sequence) VALUES($1,'invalid', $2,999)`, f.battle.ID, f.abilityID)
	requireConstraint(t, err, "23503", "")
	if _, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ query, constraint string }{
		{`INSERT INTO battle_log_archives(battle_id,final_sequence) VALUES($1,1)`, ""},
		{`INSERT INTO battle_log_archives(battle_id,final_sequence,cleared_at) VALUES($1,2,'infinity')`, "battle_log_archives_cleared_at_check"},
		{`INSERT INTO battle_log_archives(battle_id,final_sequence,cleared_at) VALUES($1,2,'2000-01-01')`, ""},
	} {
		_, err := f.s.pool.Exec(f.ctx, tc.query, f.battle.ID)
		requireConstraint(t, err, "23514", tc.constraint)
	}
	mustExec(t, f.ctx, f.s, `INSERT INTO battle_log_archives(battle_id,final_sequence) VALUES($1,2)`, f.battle.ID)
	_, err = f.s.pool.Exec(f.ctx, `DELETE FROM battle_events WHERE battle_id=$1 AND sequence=2`, f.battle.ID)
	requireConstraint(t, err, "23503", "battle_log_archives_battle_id_final_sequence_fkey")
	_, err = f.s.pool.Exec(f.ctx, `INSERT INTO battle_commands(battle_id,command_key,ability_id,event_sequence) VALUES($1,'late', $2,2)`, f.battle.ID, f.abilityID)
	requireConstraint(t, err, "23514", "")
	for _, query := range []string{`UPDATE battle_log_archives SET cleared_at=clock_timestamp() WHERE battle_id=$1`, `DELETE FROM battle_log_archives WHERE battle_id=$1`} {
		_, err := f.s.pool.Exec(f.ctx, query, f.battle.ID)
		requireConstraint(t, err, "23514", "")
	}
}
