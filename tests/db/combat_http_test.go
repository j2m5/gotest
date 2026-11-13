package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
)

func TestCombatHTTPWithAutonomousWorker(t *testing.T) {
	s, ctx, _, _, in := combatFixture(t)
	token := strings.Repeat("a", 64)
	repositoryActor(t, ctx, s, 1, auth.TokenDigest(token))
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: token}
	mustExec(t, ctx, s, `INSERT INTO character_abilities(character_id,ability_id)
 SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id ON CONFLICT DO NOTHING`)
	var abilityID int
	if err := s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=$1 ORDER BY ability_id LIMIT 1`, in.CharacterID).Scan(&abilityID); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE abilities SET damage=100000,mp_cost=10 WHERE id=$1`, abilityID)
	a := auth.NewService(s.AuthRepository, auth.Options{})
	r := db.NewCombatRepository(s.pool)
	h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(character.NewService(db.NewCharacterRepository(s.pool))), middleware.NewMiddleware(a), routes.Options{
		StaticDir: t.TempDir(), Combat: handlers.NewCombatHandler(combat.NewService(r)),
	})
	if err != nil {
		t.Fatal(err)
	}
	startBody, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	flowRequest(t, h, "POST", "/api/battles/start", string(startBody), nil, 401)
	w := flowRequest(t, h, "POST", "/api/battles/start", string(startBody), cookie, 200)
	var started combat.Result
	if err = json.Unmarshal(w.Body.Bytes(), &started); err != nil || started.Battle == nil || started.Replayed || started.EventSequence != 1 {
		t.Fatalf("start: %s, %v", w.Body, err)
	}
	battleID := started.Battle.ID
	mustExec(t, ctx, s, `UPDATE battle_executions SET next_action_at=clock_timestamp() WHERE battle_id=$1`, battleID)
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); combat.NewWorker(r, nil).Run(workerCtx) }()
	t.Cleanup(func() { cancel(); <-done })

	// Observe SQL only: no HTTP request or direct ProcessDue call may drive the NPC.
	deadline := time.After(5 * time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		var count int
		if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM battle_events WHERE battle_id=$1 AND payload->>'kind'='npc_ability'`, battleID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("NPC did not act autonomously")
		case <-tick.C:
		}
	}
	cancel()
	<-done
	path := fmt.Sprintf("/api/battles/detail?battle_id=%d&character_id=%d", battleID, in.CharacterID)
	w = flowRequest(t, h, "GET", path, "", cookie, 200)
	var view combat.View
	if err = json.Unmarshal(w.Body.Bytes(), &view); err != nil || len(view.Events) < 2 || view.Events[1].Kind != "npc_ability" || view.Events[1].Damage <= 0 {
		t.Fatalf("autonomous attack: %s, %v", w.Body, err)
	}
	body := fmt.Sprintf(`{"battle_id":"%d","character_id":%d,"ability_id":%d,"request_key":"http:win"}`, battleID, in.CharacterID, abilityID)
	for i := 0; i < 2; i++ {
		w = flowRequest(t, h, "POST", "/api/battles/use", body, cookie, 200)
		var won combat.Result
		if err = json.Unmarshal(w.Body.Bytes(), &won); err != nil || won.Battle.Status != combat.Won || won.Replayed != (i == 1) {
			t.Fatalf("victory/retry: %s, %v", w.Body, err)
		}
	}
	w = flowRequest(t, h, "GET", path, "", cookie, 200)
	if err = json.Unmarshal(w.Body.Bytes(), &view); err != nil || view.NextActionAt != nil || view.Events[len(view.Events)-1].Kind != "won" || view.Events[len(view.Events)-1].Credits <= 0 {
		t.Fatalf("victory journal: %s, %v", w.Body, err)
	}
	w = flowRequest(t, h, "GET", fmt.Sprintf("/api/battles/current?character_id=%d", in.CharacterID), "", cookie, 200)
	var current combat.State
	if err = json.Unmarshal(w.Body.Bytes(), &current); err != nil || current.Battle != nil {
		t.Fatalf("completed battle still active: %s, %v", w.Body, err)
	}
	w = flowRequest(t, h, "POST", "/api/battles/retreat", fmt.Sprintf(`{"battle_id":"%d","character_id":%d}`, battleID, in.CharacterID), cookie, 200)
	var late combat.Result
	if err = json.Unmarshal(w.Body.Bytes(), &late); err != nil || late.Battle.Status != combat.Won || !late.Replayed {
		t.Fatalf("late retreat: %s, %v", w.Body, err)
	}
}
