package db_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/combat"
	"gotest/internal/maintenance"
	"gotest/internal/middleware"
)

func TestCombatLogArchiveViewKeepsOutcomeForEveryCursor(t *testing.T) {
	f := wonLogFixture(t)
	old := cloneFinishedLog(t, f, 40)
	h, _ := characterFlowRouter(t, f.s, &flowMail{})
	token := strings.Repeat("b", 64)
	repositoryActor(t, f.ctx, f.s, f.actor.UserID, auth.TokenDigest(token))
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: token}
	url := fmt.Sprintf("/api/battles/detail?battle_id=%d&character_id=%d", old.battle.ID, old.in.CharacterID)
	read := func(cursor int) (map[string]json.RawMessage, combat.View) {
		t.Helper()
		w := flowRequest(t, h, "GET", fmt.Sprintf("%s&after_sequence=%d", url, cursor), "", cookie, 200)
		var fields map[string]json.RawMessage
		var view combat.View
		if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
			t.Fatal(err)
		}
		return fields, view
	}
	fields, before := read(0)
	if string(fields["log_cleared_at"]) != "null" {
		t.Fatal("unarchived API marker missing", string(fields["log_cleared_at"]))
	}
	if _, err := maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true}); err != nil {
		t.Fatal(err)
	}
	for _, cursor := range []int{0, 2, 3, 999} {
		fields, after := read(cursor)
		if len(fields["log_cleared_at"]) == 0 || string(fields["log_cleared_at"]) == "null" || after.NextAfter != nil || len(after.Events) != 1 || after.Events[0] != before.Events[2] {
			t.Fatal("archived result incomplete", cursor, after, string(fields["log_cleared_at"]))
		}
	}
	flowRequest(t, h, "GET", url, "", nil, 401)
}
