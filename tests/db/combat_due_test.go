package db_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"gotest/internal/combat"
	"gotest/internal/db"
)

func TestCombatDueInvalidBattleDoesNotStarveHealthyBattle(t *testing.T) {
	f := newExecutionFixture(t)
	other := f.in
	other.CharacterID, other.Key = 3, "healthy"
	mustExec(t, f.ctx, f.s, `UPDATE characters SET location_id=$1 WHERE id=$2`, other.LocationID, other.CharacterID)
	if err := f.s.pool.QueryRow(f.ctx, `SELECT s.id FROM npc_spawns s
 JOIN location_npcs n ON n.id=s.location_npc_id WHERE n.location_id=$1 AND s.id<>$2
 ORDER BY s.id LIMIT 1`, other.LocationID, f.in.NPCSpawnID).Scan(&other.NPCSpawnID); err != nil {
		t.Fatal(err)
	}
	healthy, err := f.r.Start(f.ctx, f.other, other)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, f.ctx, f.s, `UPDATE battle_executions SET next_action_at=statement_timestamp(),
 pattern_step=CASE WHEN battle_id=$1 THEN 2147483647 ELSE 0 END`, f.battle.ID)
	if _, err = f.r.ProcessDue(f.ctx); err == nil {
		t.Fatal("invalid battle did not fail")
	}
	// A fresh repository must skip the broken encounter using shared DB state.
	if worked, err := db.NewCombatRepository(f.s.pool).ProcessDue(f.ctx); err != nil || !worked {
		t.Fatal("invalid battle starved healthy encounter", worked, err)
	}
	var badEvents, goodEvents int
	var status combat.Status
	err = f.s.pool.QueryRow(f.ctx, `SELECT
 (SELECT count(*) FROM battle_events WHERE battle_id=$1),
 (SELECT count(*) FROM battle_events WHERE battle_id=$2),
 (SELECT status FROM battles WHERE id=$1)`, f.battle.ID, healthy.Battle.ID).Scan(&badEvents, &goodEvents, &status)
	if err != nil || badEvents != 1 || goodEvents < 2 || status != combat.Active {
		t.Fatal("partial failure or missing healthy attack", badEvents, goodEvents, status, err)
	}
}

func failCombatEventInsert(t *testing.T, f executionFixture) {
	t.Helper()
	mustExec(t, f.ctx, f.s, `CREATE FUNCTION fail_test_combat_event() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN RAISE EXCEPTION 'injected event write failure' USING ERRCODE='P0001'; END; $$;
 CREATE TRIGGER fail_test_combat_event BEFORE INSERT ON battle_events FOR EACH ROW EXECUTE FUNCTION fail_test_combat_event()`)
}

func combatEffectsSnapshot(t *testing.T, f executionFixture) string {
	t.Helper()
	var snapshot string
	err := f.s.pool.QueryRow(f.ctx, `SELECT jsonb_build_array(
 (SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM characters c),
 (SELECT jsonb_agg(to_jsonb(b) ORDER BY id) FROM battles b),
 (SELECT jsonb_agg(to_jsonb(e)-'retry_count'-'retry_at'-'updated_at' ORDER BY battle_id) FROM battle_executions e),
 (SELECT jsonb_agg(to_jsonb(e) ORDER BY battle_id,sequence) FROM battle_events e),
 (SELECT jsonb_agg(to_jsonb(s) ORDER BY id) FROM npc_spawns s),
 (SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM wallets w),
 (SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM wallet_entries w))::text`).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestCombatDueSQLFailureRollsBackAndRetriesWithCappedDelay(t *testing.T) {
	f := newExecutionFixture(t)
	f.due(t)
	failCombatEventInsert(t, f)
	before := combatEffectsSnapshot(t, f)
	var scheduled time.Time
	if err := f.s.pool.QueryRow(f.ctx, `SELECT next_action_at FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&scheduled); err != nil {
		t.Fatal(err)
	}
	notified := 0
	r := f.r.WithNotifications(func(int) { notified++ })
	for attempt := 1; attempt <= 8; attempt++ {
		var started, ended time.Time
		if err := f.s.pool.QueryRow(f.ctx, `SELECT clock_timestamp()`).Scan(&started); err != nil {
			t.Fatal(err)
		}
		worked, err := r.ProcessDue(f.ctx)
		var deferred *combat.ProcessingError
		var sqlErr *pgconn.PgError
		if !worked || !errors.As(err, &deferred) || !errors.As(err, &sqlErr) || sqlErr.Code != "P0001" {
			t.Fatal("failure was not isolated", worked, err)
		}
		var count int
		var retry time.Time
		err = f.s.pool.QueryRow(f.ctx, `SELECT retry_count,retry_at,clock_timestamp() FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&count, &retry, &ended)
		delay := time.Duration(min(30, 1<<(attempt-1))) * time.Second
		if err != nil || count != min(attempt, 6) || retry.Before(started.Add(delay)) || retry.After(ended.Add(delay)) {
			t.Fatal("retry delay or cap", attempt, count, retry, started, ended, err)
		}
		if deferred.BattleID != f.battle.ID || deferred.CharacterID != f.in.CharacterID || !deferred.RetryAt.Equal(retry) {
			t.Fatal("missing failure context", deferred)
		}
		if combatEffectsSnapshot(t, f) != before || notified != 0 {
			t.Fatal("partial effects or notification escaped failed transaction")
		}
		// No in-memory state: a fresh repository observes the same delay.
		if worked, err = db.NewCombatRepository(f.s.pool).ProcessDue(f.ctx); err != nil || worked {
			t.Fatal("retry ran before its deadline", worked, err)
		}
		mustExec(t, f.ctx, f.s, `UPDATE battle_executions SET retry_at=clock_timestamp()-interval '1 microsecond' WHERE battle_id=$1`, f.battle.ID)
	}
	mustExec(t, f.ctx, f.s, `DROP TRIGGER fail_test_combat_event ON battle_events`)
	if worked, err := r.ProcessDue(f.ctx); err != nil || !worked {
		t.Fatal("repaired battle did not resume", worked, err)
	}
	var clear bool
	var firstAttack time.Time
	if err := f.s.pool.QueryRow(f.ctx, `SELECT retry_count=0 AND retry_at IS NULL,
 (SELECT occurred_at FROM battle_events WHERE battle_id=$1 AND sequence=2)
 FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&clear, &firstAttack); err != nil || !clear || !firstAttack.Equal(scheduled) || notified != 1 {
		t.Fatal("recovery changed original clock or retained backoff", clear, firstAttack, scheduled, notified, err)
	}
}

func TestCombatDueConcurrentFailuresScheduleOneRetry(t *testing.T) {
	f := newExecutionFixture(t)
	f.due(t)
	failCombatEventInsert(t, f)
	start := make(chan struct{})
	results := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			<-start
			worked, err := db.NewCombatRepository(f.s.pool).ProcessDue(f.ctx)
			if worked != (err != nil) {
				results <- fmt.Errorf("inconsistent claim: worked=%v err=%w", worked, err)
				return
			}
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	failures := 0
	for err := range results {
		if err != nil {
			var deferred *combat.ProcessingError
			if !errors.As(err, &deferred) {
				t.Fatal(err)
			}
			failures++
		}
	}
	var count int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT retry_count FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&count); err != nil || count != 1 || failures != 1 {
		t.Fatal("same failed attempt claimed repeatedly", count, failures, err)
	}
}

func TestCombatDueRecoveredPlayerCommandClearsBackoff(t *testing.T) {
	for _, command := range []string{"use", "retreat", "win"} {
		t.Run(command, func(t *testing.T) {
			f := newExecutionFixture(t)
			f.due(t)
			failCombatEventInsert(t, f)
			if worked, err := f.r.ProcessDue(f.ctx); err == nil || !worked {
				t.Fatal("failed to schedule retry", worked, err)
			}
			mustExec(t, f.ctx, f.s, `UPDATE battle_executions SET retry_at=clock_timestamp()+interval '30 seconds' WHERE battle_id=$1`, f.battle.ID)
			mustExec(t, f.ctx, f.s, `DROP TRIGGER fail_test_combat_event ON battle_events`)
			var result *combat.Result
			var err error
			if command == "retreat" {
				result, err = f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID)
			} else {
				if command == "win" {
					mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
				}
				result, err = f.cast("recovered")
			}
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]combat.Status{"use": combat.Active, "retreat": combat.Retreated, "win": combat.Won}[command]
			var clear bool
			if err := f.s.pool.QueryRow(f.ctx, `SELECT retry_count=0 AND retry_at IS NULL FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&clear); err != nil || !clear || result.Battle.Status != want {
				t.Fatal("successful command retained retry state", clear, result.Battle.Status, err)
			}
		})
	}
}

func TestCombatDueFailedDeferralRollsBackAndReportsFailure(t *testing.T) {
	f := newExecutionFixture(t)
	f.due(t)
	failCombatEventInsert(t, f)
	mustExec(t, f.ctx, f.s, `CREATE FUNCTION fail_test_combat_retry() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN IF NEW.retry_count>0 THEN RAISE EXCEPTION 'injected retry write failure'; END IF; RETURN NEW; END; $$;
 CREATE TRIGGER fail_test_combat_retry BEFORE UPDATE ON battle_executions FOR EACH ROW EXECUTE FUNCTION fail_test_combat_retry()`)
	before := combatEffectsSnapshot(t, f)
	if worked, err := f.r.ProcessDue(f.ctx); err == nil || worked {
		t.Fatal("failed deferral reported successful work", worked, err)
	}
	var clear bool
	if err := f.s.pool.QueryRow(f.ctx, `SELECT retry_count=0 AND retry_at IS NULL FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&clear); err != nil || !clear || combatEffectsSnapshot(t, f) != before {
		t.Fatal("failed deferral changed battle", clear, err)
	}
}
