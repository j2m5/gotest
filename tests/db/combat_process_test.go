package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/combat"
	"gotest/internal/middleware"
)

func isolatedCombatDSN(t *testing.T, s *testStorage) string {
	t.Helper()
	schema := s.pool.Config().ConnConfig.RuntimeParams["search_path"]
	if !strings.HasPrefix(schema, "auth_test_") || strings.ContainsAny(schema, " ,;'\\") {
		t.Fatal("process test requires a single isolated fixture schema")
	}
	dsn := s.pool.Config().ConnString()
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("cannot construct isolated process configuration")
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	return dsn + " search_path=" + schema
}

func combatProcessRequest(ctx context.Context, app *pathApp, cookie *http.Cookie, method, path, body string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, app.url+path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.AddCookie(cookie)
	request.Header.Set("X-Account-ID", "1")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("combat request returned status %d", response.StatusCode)
	}
	return io.ReadAll(io.LimitReader(response.Body, 1<<20))
}

func TestCombatSurvivesActualProcessCrash(t *testing.T) {
	s, setup, _, _, in := combatFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	token := strings.Repeat("c", 64)
	repositoryActor(t, setup, s, 1, auth.TokenDigest(token))
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: token}
	mustExec(t, setup, s, `INSERT INTO character_abilities(character_id,ability_id)
 SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id WHERE c.id=2`)
	mustExec(t, setup, s, `UPDATE characters SET hp=1000,hp_current=1000,power=0,mastery=0 WHERE id=2;
 UPDATE abilities SET damage=5,mp_cost=10,cooldown_ms=6000`)
	var abilityID int
	if err := s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=2`).Scan(&abilityID); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "combat-app.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/app")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build app: %v\n%s", err, output)
	}
	port, cwd := pathAppPort(t), t.TempDir()
	values := map[string]string{"DATABASE_URL": isolatedCombatDSN(t, s), "APP_PORT": port, "PUBLIC_URL": "http://127.0.0.1:" + port, "MAIL_DRIVER": "local"}
	start := func() (*pathApp, context.CancelFunc) {
		processContext, kill := context.WithCancel(ctx)
		t.Cleanup(kill)
		app := startPathApp(t, processContext, binary, cwd, port, values, "-env-file=")
		waitPathApp(t, ctx, app)
		return app, func() { kill(); <-app.done }
	}
	first, killFirst := start()
	blocker, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback(ctx)
	var blockerPID int
	if err = blocker.QueryRow(ctx, `SELECT pg_backend_pid() FROM npc_spawns WHERE id=$1 FOR UPDATE`, in.NPCSpawnID).Scan(&blockerPID); err != nil {
		t.Fatal(err)
	}
	startBody, _ := json.Marshal(in)
	before := characterStatsSnapshot(t, ctx, s)
	waiting := make(chan error, 1)
	go func() {
		_, err := combatProcessRequest(ctx, first, cookie, "POST", "/api/battles/start", string(startBody))
		waiting <- err
	}()
	awaitResetUserLock(t, ctx, s, blockerPID, waiting)
	killFirst() // OS process termination, not a graceful worker shutdown.
	if err = <-waiting; err == nil {
		t.Fatal("queued request unexpectedly completed before process death")
	}
	if err = blocker.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var battles int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM battles`).Scan(&battles); err != nil || battles != 0 {
		t.Fatalf("crash left a partial battle: %d %v", battles, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("queued crash changed character resources")
	}

	second, killSecond := start()
	call := func(app *pathApp, method, path, body string, out any) {
		t.Helper()
		data, err := combatProcessRequest(ctx, app, cookie, method, path, body)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, out); err != nil {
			t.Fatal(err)
		}
	}
	var started, applied combat.Result
	call(second, "POST", "/api/battles/start", string(startBody), &started)
	if started.Battle == nil || started.Replayed {
		t.Fatal("rolled-back start was not available for retry")
	}
	id := started.Battle.ID
	useBody := fmt.Sprintf(`{"battle_id":"%d","character_id":2,"ability_id":%d,"request_key":"process:one"}`, id, abilityID)
	call(second, "POST", "/api/battles/use", useBody, &applied)
	if applied.Replayed || applied.Character.MPCurrent != 90 {
		t.Fatal("initial ability was not applied once")
	}
	killSecond()

	var nextAction, personal, global time.Time
	var prefix string
	var sequence int
	if err = s.pool.QueryRow(ctx, `SELECT e.next_action_at,e.event_sequence,c.global_cooldown_ready_at,ca.ready_at,
 (SELECT jsonb_agg(to_jsonb(ev) ORDER BY sequence)::text FROM battle_events ev WHERE battle_id=$1)
 FROM battle_executions e JOIN battles b ON b.id=e.battle_id JOIN characters c ON c.id=b.character_id
 JOIN character_abilities ca ON ca.character_id=c.id AND ca.ability_id=$2 WHERE e.battle_id=$1`, id, abilityID).Scan(&nextAction, &sequence, &global, &personal, &prefix); err != nil {
		t.Fatal(err)
	}
	// Allow the original NPC deadline to pass with no application process running.
	if pause := time.Until(nextAction.Add(100 * time.Millisecond)); pause > 0 {
		if pause > 4*time.Second {
			t.Fatal("unexpected next-action interval")
		}
		time.Sleep(pause)
	}
	var unchanged int
	if err = s.pool.QueryRow(ctx, `SELECT event_sequence FROM battle_executions WHERE battle_id=$1`, id).Scan(&unchanged); err != nil || unchanged != sequence {
		t.Fatal("events changed without a running process", unchanged, err)
	}
	third, _ := start()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err = s.pool.QueryRow(ctx, `SELECT event_sequence FROM battle_executions WHERE battle_id=$1`, id).Scan(&unchanged); err != nil {
			t.Fatal(err)
		}
		if unchanged > sequence {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("restarted worker did not catch up without HTTP reads")
		}
		time.Sleep(10 * time.Millisecond)
	}
	var scheduled time.Time
	var kind, savedPrefix string
	if err = s.pool.QueryRow(ctx, `SELECT occurred_at,payload->>'kind' FROM battle_events WHERE battle_id=$1 AND sequence=$2`, id, sequence+1).Scan(&scheduled, &kind); err != nil {
		t.Fatal(err)
	}
	if kind != "npc_ability" || !scheduled.Equal(nextAction) {
		t.Fatalf("restart moved the saved deadline: %s %s vs %s", kind, scheduled, nextAction)
	}
	if err = s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(ev) ORDER BY sequence)::text FROM battle_events ev WHERE battle_id=$1 AND sequence<=$2`, id, sequence).Scan(&savedPrefix); err != nil || prefix != savedPrefix {
		t.Fatal("restart rewrote existing history", err)
	}
	var replayed combat.Result
	call(third, "POST", "/api/battles/use", useBody, &replayed)
	if !replayed.Replayed || replayed.EventSequence != applied.EventSequence || replayed.Battle.ID != id {
		t.Fatal("command receipt did not survive restart")
	}
	var savedPersonal, savedGlobal time.Time
	var casts, receipts int
	if err = s.pool.QueryRow(ctx, `SELECT ca.ready_at,c.global_cooldown_ready_at,
 (SELECT count(*) FROM battle_events WHERE battle_id=$1 AND payload->>'kind'='player_ability'),
 (SELECT count(*) FROM battle_commands WHERE battle_id=$1) FROM characters c
 JOIN character_abilities ca ON ca.character_id=c.id AND ca.ability_id=$2 WHERE c.id=2`, id, abilityID).Scan(&savedPersonal, &savedGlobal, &casts, &receipts); err != nil {
		t.Fatal(err)
	}
	if !savedPersonal.Equal(personal) || !savedGlobal.Equal(global) || casts != 1 || receipts != 1 {
		t.Fatal("restart/replay reset cooldown or duplicated action")
	}
	beforeEnergy, afterEnergy := applied.Character.Regeneration, replayed.Character.Regeneration
	if beforeEnergy.MPBaseline != afterEnergy.MPBaseline || beforeEnergy.MPRemainder != afterEnergy.MPRemainder || !beforeEnergy.MPUpdatedAt.Equal(afterEnergy.MPUpdatedAt) {
		t.Fatal("replay changed the persisted energy baseline; regenerated current MP cannot hide another debit")
	}
	var ended combat.Result
	call(third, "POST", "/api/battles/retreat", fmt.Sprintf(`{"battle_id":"%d","character_id":2}`, id), &ended)
	if ended.Battle.Status != combat.Retreated || ended.Battle.NPC.HPCurrent != ended.Battle.NPC.HP {
		t.Fatal("recovered encounter cannot retreat")
	}
	t.Log("PASS actual process death rolls back queued start; restart preserves history, NPC deadline, cooldowns and command receipt")
}
