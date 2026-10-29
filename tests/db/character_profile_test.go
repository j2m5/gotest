package db_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/middleware"
)

func TestPublicCharacterProfileKeepsAuthorIdentityAndPrivateData(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	token := strings.Repeat("a", 64)
	if err := createStoredSession(t, s, ctx, 2, auth.TokenDigest(token), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: token}
	read := func() map[string]any {
		t.Helper()
		w := flowRequest(t, h, "GET", "/api/characters/profile?character_id=1", "", cookie, 200)
		var envelope map[string]map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if len(envelope) != 1 || envelope["character"] == nil {
			t.Fatalf("unexpected envelope: %s", w.Body)
		}
		profile := envelope["character"]
		keys := make([]string, 0, len(profile))
		for key := range profile {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		want := []string{"account_role", "avatar", "class_id", "faction_id", "id", "level", "name"}
		if !reflect.DeepEqual(keys, want) {
			t.Fatalf("profile exposes unexpected fields: %v", keys)
		}
		if profile["id"] != float64(1) || profile["name"] != "First character" || profile["level"] != float64(1) {
			t.Fatalf("wrong character: %+v", profile)
		}
		return profile
	}
	first := read()
	if first["account_role"] != "user" || first["avatar"] != nil || activeCharacter(t, ctx, s, 2) != nil {
		t.Fatalf("profile without viewer selection: %+v", first)
	}
	// Switching the author's active character cannot change the historical link.
	mustExec(t, ctx, s, `UPDATE users SET character_id=2,role_id=2 WHERE id=1;
		UPDATE characters SET avatar='images/avatars/sith_warrior.gif' WHERE id=1`)
	second := read()
	if second["account_role"] != "moderator" || second["avatar"] != "images/avatars/sith_warrior.gif" || *activeCharacter(t, ctx, s, 1) != 2 {
		t.Fatalf("profile after author switch: %+v", second)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=3 WHERE id=2; UPDATE users SET role_id=3 WHERE id=1`)
	if read()["account_role"] != "administrator" || *activeCharacter(t, ctx, s, 2) != 3 {
		t.Fatal("viewing another profile changed the viewer or lost the current account role")
	}
}

func TestPublicCharacterProfileRepositoryRechecksSession(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	repo := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 2, "profile-session")
	if _, err := repo.Profile(ctx, actor, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Profile(ctx, character.Actor{UserID: 1, SessionHash: actor.SessionHash}, 1); !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("mismatched viewer identity: %v", err)
	}
	mustExec(t, ctx, s, `DELETE FROM sessions WHERE user_id=2`)
	if _, err := repo.Profile(ctx, actor, 1); !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("revoked session: %v", err)
	}
}

func TestPublicCharacterProfileRequiresSessionAndValidID(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	url := "/api/characters/profile?character_id=1"
	flowRequest(t, h, "GET", url, "", nil, 401)
	token := strings.Repeat("b", 64)
	if err := createStoredSession(t, s, ctx, 2, auth.TokenDigest(token), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: token}
	for _, id := range []string{"", "0", "-1", "abc", "1.5", "2147483648", "999999999999999999999999"} {
		flowRequest(t, h, "GET", "/api/characters/profile?character_id="+id, "", cookie, 400)
	}
	flowRequest(t, h, "GET", "/api/characters/profile?character_id=2147483647", "", cookie, 404)
	flowRequest(t, h, "POST", url, "{}", cookie, 405)
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=2`)
	flowRequest(t, h, "GET", url, "", cookie, 401)
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()+interval '1 hour' WHERE user_id=2; UPDATE users SET email_verified_at=NULL WHERE id=2`)
	flowRequest(t, h, "GET", url, "", cookie, 401)
	mustExec(t, ctx, s, `UPDATE users SET email_verified_at=clock_timestamp() WHERE id=2;
		INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(2,'account','Restriction',1)`)
	flowRequest(t, h, "GET", url, "", cookie, 401)
	mustExec(t, ctx, s, `DELETE FROM user_bans WHERE user_id=2; DELETE FROM sessions WHERE user_id=2`)
	flowRequest(t, h, "GET", url, "", cookie, 401)
}
