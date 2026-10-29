package db_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/middleware"
	"gotest/internal/presence"
)

func presenceCookie(t *testing.T, s *testStorage, ctx context.Context, userID int, seed string) *http.Cookie {
	t.Helper()
	token := strings.Repeat(seed, 64)
	if err := createStoredSession(t, s, ctx, userID, auth.TokenDigest(token), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: middleware.SessionCookie, Value: token}
}

func presenceState(t *testing.T, h http.Handler, cookie *http.Cookie) (int, map[int]bool) {
	t.Helper()
	w := flowRequest(t, h, "GET", "/api/presence?character_ids=1,2,3,999", "", cookie, 200)
	var result struct {
		Total      int `json:"total"`
		Characters []struct {
			ID     int  `json:"character_id"`
			Online bool `json:"is_online"`
		} `json:"characters"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	states := map[int]bool{}
	for _, c := range result.Characters {
		states[c.ID] = c.Online
	}
	if len(states) != 4 {
		t.Fatalf("incomplete statuses: %s", w.Body)
	}
	return result.Total, states
}

func TestCharacterPresenceExpiresWithoutActionsDespitePolling(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	cookie := presenceCookie(t, s, ctx, 1, "a")
	if total, states := presenceState(t, h, cookie); total != 0 || states[1] {
		t.Fatal("login alone marked a character online")
	}
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":1}`, cookie, 200)
	if total, states := presenceState(t, h, cookie); total != 1 || !states[1] || states[2] || states[999] {
		t.Fatalf("selection: %d %v", total, states)
	}
	mustExec(t, ctx, s, `UPDATE sessions SET last_seen_at=clock_timestamp()-interval '121 seconds' WHERE user_id=1`)
	var before, after time.Time
	if err := s.pool.QueryRow(ctx, `SELECT last_seen_at FROM sessions WHERE user_id=1`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		flowRequest(t, h, "GET", "/api/auth/me", "", cookie, 200)
		flowRequest(t, h, "GET", "/api/characters", "", cookie, 200)
		if total, states := presenceState(t, h, cookie); total != 0 || states[1] {
			t.Fatalf("idle polling extended presence: %d %v", total, states)
		}
	}
	if err := s.pool.QueryRow(ctx, `SELECT last_seen_at FROM sessions WHERE user_id=1`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if !before.Equal(after) {
		t.Fatal("GET changed the activity timestamp")
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, cookie, 200)
	if total, states := presenceState(t, h, cookie); total != 1 || !states[1] {
		t.Fatal("new action did not restore presence")
	}
	if err := s.pool.QueryRow(ctx, `SELECT last_seen_at FROM sessions WHERE user_id=1`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, cookie, 200)
	if err := s.pool.QueryRow(ctx, `SELECT last_seen_at FROM sessions WHERE user_id=1`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if !before.Equal(after) {
		t.Fatal("unthrottled repeated presence write")
	}
}

func TestCharacterPresenceRejectsStaleSelectionAndLogout(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	one, viewer := presenceCookie(t, s, ctx, 1, "b"), presenceCookie(t, s, ctx, 2, "c")
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":1}`, one, 200)
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":2}`, one, 200)
	w := flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, one, 409)
	if !strings.Contains(w.Body.String(), "character_changed") {
		t.Fatal(w.Body)
	}
	if total, states := presenceState(t, h, viewer); total != 1 || states[1] || !states[2] {
		t.Fatalf("stale character revived: %d %v", total, states)
	}
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":null}`, one, 200)
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":2}`, one, 409)
	if total, _ := presenceState(t, h, viewer); total != 0 {
		t.Fatal("cleared character still online")
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":null}`, one, 200)
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":1}`, one, 200)
	flowRequest(t, h, "POST", "/api/auth/logout", "", one, 200)
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, one, 401)
	if total, _ := presenceState(t, h, viewer); total != 0 {
		t.Fatal("logout did not remove presence")
	}
}

func TestCharacterPresenceDeduplicatesAndRequiresLiveUnbannedSession(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	one, two, viewer := presenceCookie(t, s, ctx, 1, "d"), presenceCookie(t, s, ctx, 1, "e"), presenceCookie(t, s, ctx, 2, "f")
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":1}`, one, 200)
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, two, 200)
	if total, _ := presenceState(t, h, viewer); total != 1 {
		t.Fatal("sessions counted as separate characters")
	}
	mustExec(t, ctx, s, `DELETE FROM sessions WHERE token=$1`, auth.TokenDigest(one.Value))
	if total, _ := presenceState(t, h, viewer); total != 1 {
		t.Fatal("revocation removed another live session")
	}
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
	if total, _ := presenceState(t, h, viewer); total != 0 {
		t.Fatal("expired session counted online")
	}
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()+interval '1 hour' WHERE user_id=1;
		INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,created_at,expires_at) VALUES(1,'chat','chat only',2,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP+interval '1 hour')`)
	if total, _ := presenceState(t, h, viewer); total != 1 {
		t.Fatal("chat ban hid player presence")
	}
	mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','account ban',2)`)
	if total, _ := presenceState(t, h, viewer); total != 0 {
		t.Fatal("account banned player counted online")
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":1}`, two, 401)
}

func TestCharacterPresenceHTTPValidationAndAccountIsolation(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	cookie := presenceCookie(t, s, ctx, 1, "7")
	for _, path := range []string{"/api/presence", "/api/presence/online"} {
		flowRequest(t, h, "GET", path, "", nil, 401)
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":null}`, nil, 401)
	for _, suffix := range []string{"0", "-1", "1.5", "2147483648", "1,,2", strings.Repeat("1,", 100) + "2"} {
		flowRequest(t, h, "GET", "/api/presence?character_ids="+suffix, "", cookie, 400)
	}
	for _, body := range []string{`{}`, `{"character_id":"1"}`, `{"character_id":0}`, `{"character_id":-1}`, `{"character_id":2147483648}`, `{"character_id":1,"user_id":2}`} {
		flowRequest(t, h, "POST", "/api/presence/heartbeat", body, cookie, 400)
	}
	flowRequest(t, h, "POST", "/api/presence/heartbeat", `{"character_id":3}`, cookie, 409)
	flowRequest(t, h, "GET", "/api/presence/online?after_id=-1", "", cookie, 400)
	flowRequest(t, h, "PUT", "/api/presence", "{}", cookie, 405)
	r := httptest.NewRequest("POST", "/api/presence/heartbeat", strings.NewReader(`{"character_id":null}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Account-ID", "2")
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 409 || !strings.Contains(w.Body.String(), "account_changed") {
		t.Fatalf("cross-account signal: %d %s", w.Code, w.Body)
	}
}

func TestCharacterPresenceDirectoryPaginationAndPrivacy(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	viewer := presenceCookie(t, s, ctx, 1, "8")
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at)
		SELECT 'online'||n||'@example.test','online'||n,'unused',now() FROM generate_series(10,60) n;
		INSERT INTO characters(user_id,faction_id,class_id,name)
		SELECT u.id,c.faction_id,c.id,'Online player '||u.id FROM users u CROSS JOIN classes c WHERE u.login LIKE 'online%' AND c.code='trooper';
		UPDATE users u SET character_id=c.id FROM characters c WHERE c.user_id=u.id AND u.login LIKE 'online%';
		INSERT INTO sessions(user_id,token,expires_at,presence_character_id,last_seen_at)
		SELECT id,repeat(md5(id::text),2),clock_timestamp()+interval '1 hour',character_id,clock_timestamp() FROM users WHERE login LIKE 'online%'`)
	read := func(after int) presence.Directory {
		t.Helper()
		w := flowRequest(t, h, "GET", fmt.Sprintf("/api/presence/online?after_id=%d", after), "", viewer, 200)
		var page presence.Directory
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		var raw struct {
			Characters []map[string]any `json:"characters"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}
		for _, entry := range raw.Characters {
			if len(entry) != 7 || entry["name"] == nil || entry["account_role"] == nil {
				t.Fatalf("private/invalid profile fields: %v", entry)
			}
		}
		return page
	}
	first := read(0)
	if first.Total != 51 || len(first.Characters) != 50 || first.NextAfterID == nil {
		t.Fatalf("first page: %+v", first)
	}
	lastID := first.Characters[49].ID
	if *first.NextAfterID != lastID {
		t.Fatal("cursor does not match last returned character")
	}
	second := read(lastID)
	if second.Total != 51 || len(second.Characters) != 1 || second.Characters[0].ID <= lastID || second.NextAfterID != nil {
		t.Fatalf("second page: %+v", second)
	}
	mustExec(t, ctx, s, `UPDATE sessions SET last_seen_at=clock_timestamp()-interval '121 seconds' WHERE presence_character_id IS NOT NULL`)
	if empty := read(0); empty.Total != 0 || len(empty.Characters) != 0 || empty.NextAfterID != nil {
		t.Fatalf("idle directory: %+v", empty)
	}
}

func TestCharacterPresenceQueuedActivityRechecksAuthorization(t *testing.T) {
	for _, change := range []string{"session", "ban", "selection"} {
		t.Run(change, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			actor := repositoryActor(t, ctx, s, 1, "queued-presence")
			mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
			repo := db.NewPresenceRepository(s.pool)
			tx, pid := lockResetUser(t, ctx, s)
			done := make(chan error, 1)
			go func() { done <- repo.Heartbeat(ctx, actor, new(1)) }()
			awaitResetUserLock(t, ctx, s, pid, done)
			sql := `DELETE FROM sessions WHERE user_id=1`
			expected := character.ErrUnauthenticated
			if change == "ban" {
				sql = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','blocked while waiting',2)`
			}
			if change == "selection" {
				sql = `UPDATE users SET character_id=2 WHERE id=1`
				expected = presence.ErrCharacterChanged
			}
			if _, err := tx.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, expected) {
				t.Fatalf("stale %s activity: %v", change, err)
			}
			var touched int
			if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE last_seen_at IS NOT NULL`).Scan(&touched); err != nil || touched != 0 {
				t.Fatalf("unauthorized activity persisted: %d %v", touched, err)
			}
		})
	}
}

func TestCharacterPresenceMigrationPreservesSessions(t *testing.T) {
	const migration = "20261030000100_character_presence"
	s, ctx := characterTestStorageBefore(t, migration)
	_ = presenceCookie(t, s, ctx, 1, "9")
	rows := func() string {
		var result string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(s)-'presence_character_id'-'last_seen_at' ORDER BY id)::text FROM sessions s`).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	before := rows()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, migration+"."+direction+".sql")
		if rows() != before {
			t.Fatalf("%s modified existing session data", direction)
		}
		if direction == "up" {
			var seen *time.Time
			if err := s.pool.QueryRow(ctx, `SELECT last_seen_at FROM sessions LIMIT 1`).Scan(&seen); err != nil || seen != nil {
				t.Fatalf("existing session started online: %v %v", seen, err)
			}
			_, err := s.pool.Exec(ctx, `UPDATE sessions SET last_seen_at='infinity'`)
			requireConstraint(t, err, "23514", "sessions_last_seen_finite_check")
		}
	}
}
