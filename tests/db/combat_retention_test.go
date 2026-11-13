package db_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gotest/internal/combat"
	"gotest/internal/maintenance"
)

// Clone a real completed encounter into the past without disabling its guards.
// The recent original remains as a control; reward ledgers are never duplicated.
func cloneFinishedLog(t *testing.T, f executionFixture, days int) executionFixture {
	t.Helper()
	copy := f
	copy.battle = new(combat.Battle)
	*copy.battle = *f.battle
	copy.in.Key = fmt.Sprintf("old:%d:%d", f.battle.ID, days)
	err := f.s.pool.QueryRow(f.ctx, `INSERT INTO battles(character_id,npc_spawn_id,location_id,start_key,
 status,npc_template_id,npc_name,npc_level,npc_rank,npc_hp,npc_hp_current,created_at,updated_at,ended_at)
 SELECT character_id,npc_spawn_id,location_id,$3,status,npc_template_id,npc_name,npc_level,npc_rank,
 npc_hp,npc_hp_current,created_at-$2*interval '1 day',updated_at-$2*interval '1 day',ended_at-$2*interval '1 day'
 FROM battles WHERE id=$1 RETURNING id,created_at,ended_at,status`, f.battle.ID, days, copy.in.Key).
		Scan(&copy.battle.ID, &copy.battle.CreatedAt, &copy.battle.EndedAt, &copy.battle.Status)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, f.ctx, f.s, `INSERT INTO battle_executions(battle_id,rules,pattern_step,next_action_at,
 expires_at,event_sequence,created_at,updated_at)
 SELECT $3,rules,pattern_step,next_action_at-$2*interval '1 day',expires_at-$2*interval '1 day',
 event_sequence,created_at-$2*interval '1 day',updated_at-$2*interval '1 day'
 FROM battle_executions WHERE battle_id=$1`, f.battle.ID, days, copy.battle.ID)
	mustExec(t, f.ctx, f.s, `INSERT INTO battle_events(battle_id,sequence,occurred_at,payload,created_at)
 SELECT $3,sequence,occurred_at-$2*interval '1 day',
 payload||jsonb_build_object('occurred_at',occurred_at-$2*interval '1 day'),created_at-$2*interval '1 day'
 FROM battle_events WHERE battle_id=$1`, f.battle.ID, days, copy.battle.ID)
	mustExec(t, f.ctx, f.s, `INSERT INTO battle_commands(battle_id,command_key,ability_id,event_sequence,created_at)
 SELECT $3,command_key,ability_id,event_sequence,created_at-$2*interval '1 day'
 FROM battle_commands WHERE battle_id=$1`, f.battle.ID, days, copy.battle.ID)
	return copy
}

func wonLogFixture(t *testing.T) executionFixture {
	t.Helper()
	f := newExecutionFixture(t)
	mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
	result, err := f.cast("finish")
	if err != nil || result.Battle.Status != combat.Won {
		t.Fatal("fixture victory", result, err)
	}
	f.battle = result.Battle
	return f
}

func logCounts(t *testing.T, f executionFixture, events, archives int) {
	t.Helper()
	var gotEvents, gotArchives int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT
 (SELECT count(*) FROM battle_events WHERE battle_id=$1),
 (SELECT count(*) FROM battle_log_archives WHERE battle_id=$1)`, f.battle.ID).Scan(&gotEvents, &gotArchives); err != nil {
		t.Fatal(err)
	}
	if gotEvents != events || gotArchives != archives {
		t.Fatalf("battle%d events=%d archives=%d; want %d/%d", f.battle.ID, gotEvents, gotArchives, events, archives)
	}
}

func TestCombatLogRetentionPreviewAndReplay(t *testing.T) {
	recent := wonLogFixture(t)
	old := cloneFinishedLog(t, recent, 40)
	before := characterStatsSnapshot(t, old.ctx, old.s)
	protected := retainedCombatSnapshot(t, old)
	recentEvents := recent.view(t).Events
	reward := recentEvents[len(recentEvents)-1]
	options := maintenance.Options{BatchSize: 2, DryRun: true, BattleLogs: true}
	preview, err := maintenance.Run(old.ctx, old.s.pool, options)
	if err != nil || preview.BattleLogs != 1 || preview.BattleEvents != 2 || preview.Cutoff.Sub(preview.BattleLogCutoff) != 30*24*time.Hour {
		t.Fatal("preview", preview, err)
	}
	logCounts(t, old, 3, 0)
	if _, err := maintenance.Run(old.ctx, old.s.pool, maintenance.Options{BatchSize: 2}); err != nil {
		t.Fatal(err)
	}
	logCounts(t, old, 3, 0)
	options.DryRun = false
	result, err := maintenance.Run(old.ctx, old.s.pool, options)
	if err != nil || result.BattleLogs != 1 || result.BattleEvents != 2 || result.Batches != 1 {
		t.Fatal("cleanup", result, err)
	}
	logCounts(t, old, 1, 1)
	logCounts(t, recent, 3, 0)
	view := old.view(t)
	if len(view.Events) != 1 || view.Events[0].Kind != "won" || view.Events[0].Credits != reward.Credits || view.Events[0].Experience != reward.Experience {
		t.Fatal("outcome missing", view.Events)
	}
	again, err := old.cast("finish")
	if err != nil || !again.Replayed || again.EventSequence != 2 || again.Battle.Status != combat.Won {
		t.Fatal("old receipt lost", again, err)
	}
	if _, err = old.cast("new"); !errors.Is(err, combat.ErrFinished) {
		t.Fatal("closed encounter accepted new command", err)
	}
	if _, err = old.r.Use(old.ctx, old.actor, old.battle.ID, combat.UseInput{CharacterID: old.in.CharacterID, AbilityID: old.abilityID + 1, Key: "finish"}); !errors.Is(err, combat.ErrConflict) {
		t.Fatal("receipt mismatch", err)
	}
	start, err := old.r.Start(old.ctx, old.actor, old.in)
	if err != nil || !start.Replayed || start.Battle.ID != old.battle.ID || start.EventSequence != 1 {
		t.Fatal("start replay", start, err)
	}
	retreat, err := old.r.Retreat(old.ctx, old.actor, old.battle.ID, old.in.CharacterID)
	if err != nil || !retreat.Replayed || retreat.Battle.Status != combat.Won || characterStatsSnapshot(t, old.ctx, old.s) != before || retainedCombatSnapshot(t, old) != protected {
		t.Fatal("retention or replay changed resources/rewards", retreat, err)
	}
	result, err = maintenance.Run(old.ctx, old.s.pool, options)
	if err != nil || result.BattleLogs != 0 || result.BattleEvents != 0 {
		t.Fatal("second pass", result, err)
	}
}

func retainedCombatSnapshot(t *testing.T, f executionFixture) string {
	t.Helper()
	var snapshot string
	err := f.s.pool.QueryRow(f.ctx, `SELECT jsonb_build_array(
 (SELECT jsonb_agg(to_jsonb(b) ORDER BY id) FROM battles b),
 (SELECT jsonb_agg(to_jsonb(e) ORDER BY battle_id) FROM battle_executions e),
 (SELECT jsonb_agg(to_jsonb(c) ORDER BY battle_id,command_key) FROM battle_commands c),
 (SELECT jsonb_agg(to_jsonb(s) ORDER BY id) FROM npc_spawns s),
 (SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM wallets w),
 (SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM wallet_entries w))::text`).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestCombatLogRetentionAgeAndActiveBattle(t *testing.T) {
	f := wonLogFixture(t)
	beforeBoundary := cloneFinishedLog(t, f, 29)
	boundary := cloneFinishedLog(t, f, 30)
	// A separate spawn allows a current battle alongside the historical records.
	var spawn int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT s.id FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id
 WHERE n.location_id=$1 AND s.id<>$2 ORDER BY s.id LIMIT 1`, f.in.LocationID, f.in.NPCSpawnID).Scan(&spawn); err != nil {
		t.Fatal(err)
	}
	activeIn := f.in
	activeIn.Key, activeIn.NPCSpawnID = "still-active", spawn
	active, err := f.r.Start(f.ctx, f.actor, activeIn)
	if err != nil {
		t.Fatal(err)
	}
	result, err := maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
	if err != nil || result.BattleLogs != 1 {
		t.Fatal(result, err)
	}
	logCounts(t, beforeBoundary, 3, 0)
	logCounts(t, boundary, 1, 1)
	f.battle = active.Battle
	logCounts(t, f, 1, 0)
}

func TestCombatLogRetentionSkipsLocksAndConcurrentRuns(t *testing.T) {
	f := wonLogFixture(t)
	locked := cloneFinishedLog(t, f, 40)
	other := cloneFinishedLog(t, f, 41)
	tx, err := f.s.pool.Begin(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(f.ctx, `SELECT id FROM battles WHERE id=$1 FOR UPDATE`, locked.battle.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	result, err := maintenance.Run(ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
	if err != nil || result.BattleLogs != 1 {
		t.Fatal("locked encounter blocked cleanup", result, err)
	}
	logCounts(t, locked, 3, 0)
	logCounts(t, other, 1, 1)
	if err := tx.Rollback(f.ctx); err != nil {
		t.Fatal(err)
	}
	_ = cloneFinishedLog(t, f, 42)
	var wg sync.WaitGroup
	var results [2]maintenance.Result
	var errs [2]error
	for i := range 2 {
		wg.Go(func() {
			results[i], errs[i] = maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
		})
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil || results[0].BattleLogs+results[1].BattleLogs != 2 || results[0].BattleEvents+results[1].BattleEvents != 4 {
		t.Fatal("parallel cleanup counts", results, errs)
	}
	logCounts(t, locked, 1, 1)
}

func TestCombatLogRetentionRollbackKeepsCurrentLog(t *testing.T) {
	f := wonLogFixture(t)
	first := cloneFinishedLog(t, f, 41)
	failing := cloneFinishedLog(t, f, 40)
	mustExec(t, f.ctx, f.s, fmt.Sprintf(`CREATE FUNCTION fail_retention_delete() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN IF OLD.battle_id=%d THEN RAISE EXCEPTION 'injected retention failure'; END IF; RETURN OLD; END $$;
 CREATE TRIGGER zz_retention_failure BEFORE DELETE ON battle_events FOR EACH ROW EXECUTE FUNCTION fail_retention_delete()`, failing.battle.ID))
	result, err := maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
	if err == nil || result.BattleLogs != 1 || result.BattleEvents != 2 {
		t.Fatal("expected partial confirmed progress", result, err)
	}
	logCounts(t, first, 1, 1)
	logCounts(t, failing, 3, 0)
	mustExec(t, f.ctx, f.s, `DROP TRIGGER zz_retention_failure ON battle_events`)
	result, err = maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
	if err != nil || result.BattleLogs != 1 {
		t.Fatal(result, err)
	}
	logCounts(t, failing, 1, 1)
}

func TestCombatLogRetentionCancellationRollsBackArchive(t *testing.T) {
	f := cloneFinishedLog(t, wonLogFixture(t), 40)
	mustExec(t, f.ctx, f.s, `CREATE SEQUENCE retention_delete_started;
 CREATE FUNCTION wait_retention_delete() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN PERFORM nextval('retention_delete_started'); PERFORM pg_sleep(3); RETURN OLD; END $$;
 CREATE TRIGGER zz_retention_wait BEFORE DELETE ON battle_events FOR EACH ROW EXECUTE FUNCTION wait_retention_delete()`)
	ctx, cancel := context.WithCancel(f.ctx)
	done := make(chan error, 1)
	go func() {
		_, err := maintenance.Run(ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true})
		done <- err
		close(done)
	}()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var started bool
		if err := f.s.pool.QueryRow(f.ctx, `SELECT is_called FROM retention_delete_started`).Scan(&started); err != nil {
			t.Fatal(err)
		}
		if started {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cleanup never reached deletion")
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	if err := <-done; err == nil {
		t.Fatal("cancelled deletion succeeded")
	}
	logCounts(t, f, 3, 0)
}
