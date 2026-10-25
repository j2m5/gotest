package routes_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest/internal/character"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/models"
	"gotest/internal/routes"
)

type characterStub struct {
	calls  int
	actor  character.Actor
	input  character.CreateInput
	active *int
	err    error
}

func (s *characterStub) Catalog(context.Context) (*character.Catalog, error) {
	s.calls++
	return &character.Catalog{Factions: []character.Faction{{ID: 2, Code: "empire", Name: "Империя"}}, Classes: []character.Class{{ID: 8, FactionID: 2, Name: "Ситх"}}}, s.err
}
func (s *characterStub) List(_ context.Context, actor character.Actor) (*character.Roster, error) {
	s.calls++
	s.actor = actor
	return &character.Roster{Characters: []character.Character{}}, s.err
}
func (s *characterStub) Create(_ context.Context, actor character.Actor, input character.CreateInput) (*character.Character, error) {
	s.calls++
	s.actor, s.input = actor, input
	return &character.Character{ID: 10, Name: input.Name, ClassID: input.ClassID, FactionID: input.FactionID, Level: 1}, s.err
}
func (s *characterStub) SetActive(_ context.Context, actor character.Actor, id *int) (*character.Character, error) {
	s.calls++
	s.actor, s.active = actor, id
	if id == nil {
		return nil, s.err
	}
	return &character.Character{ID: *id, Name: "Малгус", Level: 1}, s.err
}
func characterRouter(t *testing.T, service *characterStub) http.Handler {
	t.Helper()
	auth := &authStub{user: &models.User{ID: 7, Login: "player", Email: "p@example.test", Password: "password-must-not-leak"}}
	h, err := routes.New(handlers.NewHandler(auth, false), handlers.NewCharacterHandler(service), middleware.NewMiddleware(auth), routes.Options{StaticDir: t.TempDir(), TrustedOrigin: "http://localhost:5173"})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func characterCookie() *http.Cookie {
	return &http.Cookie{Name: middleware.SessionCookie, Value: strings.Repeat("c", 64)}
}

func TestCharacterEndpointsRequireSession(t *testing.T) {
	s := &characterStub{}
	h := characterRouter(t, s)
	for _, tc := range []struct{ method, path, body string }{
		{"GET", "/api/characters/catalog", ""}, {"GET", "/api/characters", ""},
		{"POST", "/api/characters", `{"name":"Malgus","faction_id":2,"class_id":8}`},
		{"PUT", "/api/characters/active", `{"character_id":10}`},
	} {
		w := request(h, tc.method, tc.path, tc.body, nil)
		if w.Code != 401 || s.calls != 0 {
			t.Fatalf("unauthenticated %s: %d %s", tc.path, w.Code, w.Body)
		}
	}
}

func TestCharacterLifecycleHTTPUsesAuthenticatedOwner(t *testing.T) {
	s := &characterStub{}
	h := characterRouter(t, s)
	cookie := characterCookie()
	digest := sha256.Sum256([]byte(cookie.Value))
	wantActor := character.Actor{UserID: 7, SessionHash: hex.EncodeToString(digest[:])}
	w := request(h, "GET", "/api/characters/catalog", "", cookie)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"faction_id":2`) {
		t.Fatalf("catalog: %d %s", w.Code, w.Body)
	}
	w = request(h, "GET", "/api/characters", "", cookie)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"characters":[]`) || !strings.Contains(w.Body.String(), `"active_character_id":null`) || s.actor != wantActor {
		t.Fatalf("roster: %d %s actor=%+v", w.Code, w.Body, s.actor)
	}
	w = request(h, "POST", "/api/characters", `{"name":"Malgus","faction_id":2,"class_id":8}`, cookie)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"character":{"id":10`) || s.actor != wantActor || s.input.Name != "Malgus" {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	w = request(h, "PUT", "/api/characters/active", `{"character_id":10}`, cookie)
	if w.Code != 200 || s.active == nil || *s.active != 10 || s.actor != wantActor {
		t.Fatalf("select: %d %s", w.Code, w.Body)
	}
	w = request(h, "PUT", "/api/characters/active", `{"character_id":null}`, cookie)
	if w.Code != 200 || s.active != nil || strings.TrimSpace(w.Body.String()) != `{"character":null}` {
		t.Fatalf("clear: %d %s", w.Code, w.Body)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("selection must not be cached")
	}
	for _, secret := range []string{"password", "session", "user_id", wantActor.SessionHash} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("internal data in response: %s", secret)
		}
	}
}

func TestCharacterBodiesRejectAmbiguousOrServerOwnedFields(t *testing.T) {
	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/api/characters", `{"name":"Malgus","faction_id":2,"class_id":8,"user_id":99}`},
		{"POST", "/api/characters", `{"name":"Malgus","faction_id":2,"class_id":8,"level":50}`},
		{"PUT", "/api/characters/active", `{}`},
		{"PUT", "/api/characters/active", `{"character_id":"10"}`},
		{"PUT", "/api/characters/active", `{"character_id":1.5}`},
		{"PUT", "/api/characters/active", `{"character_id":null,"user_id":99}`},
		{"PUT", "/api/characters/active", `null`},
		{"PUT", "/api/characters/active", `{"character_id":null} {}`},
	} {
		s := &characterStub{}
		w := request(characterRouter(t, s), tc.method, tc.path, tc.body, characterCookie())
		if w.Code != 400 || s.calls != 0 {
			t.Fatalf("bad body %s: %d %s calls=%d", tc.body, w.Code, w.Body, s.calls)
		}
	}
}

func TestCharacterErrorsHaveStableStatuses(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{character.ErrUnauthenticated, 401}, {character.ErrNotFound, 404}, {character.ErrNameTaken, 409}, {character.ErrInvalidClass, 422},
		{&character.ValidationError{Messages: []string{"name invalid"}}, 422}, {errors.New("private SQL detail"), 500},
	} {
		s := &characterStub{err: tc.err}
		w := request(characterRouter(t, s), "PUT", "/api/characters/active", `{"character_id":10}`, characterCookie())
		var body map[string]any
		if w.Code != tc.status || json.Unmarshal(w.Body.Bytes(), &body) != nil || body["error"] == nil || strings.Contains(w.Body.String(), "private SQL") {
			t.Fatalf("error %v: %d %s", tc.err, w.Code, w.Body)
		}
	}
}

func TestCharacterMethodsAndCrossOriginProtection(t *testing.T) {
	s := &characterStub{}
	h := characterRouter(t, s)
	for _, tc := range []struct{ path, allow string }{{"/api/characters", "GET, POST"}, {"/api/characters/catalog", "GET"}, {"/api/characters/active", "PUT"}} {
		w := request(h, "DELETE", tc.path, "", characterCookie())
		if w.Code != 405 || w.Header().Get("Allow") != tc.allow || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("method: %d %s allow=%s", w.Code, w.Body, w.Header().Get("Allow"))
		}
	}
	r := httptest.NewRequest("PUT", "/api/characters/active", strings.NewReader(`{"character_id":null}`))
	r.AddCookie(characterCookie())
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "https://attacker.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 || s.calls != 0 {
		t.Fatalf("cross-site write: %d %s", w.Code, w.Body)
	}
}

func TestCurrentAccountIncludesNullableActiveCharacterWithoutCredentials(t *testing.T) {
	for _, id := range []*int{nil, new(42)} {
		s := &authStub{user: &models.User{ID: 7, CharacterID: id, Email: "p@example.test", Login: "player", Password: "never-expose-this"}}
		w := request(router(t, s, false, 30), "GET", "/api/auth/me", "", characterCookie())
		var body struct {
			User map[string]json.RawMessage `json:"user"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		field, ok := body.User["character_id"]
		if !ok || w.Code != 200 || strings.Contains(w.Body.String(), "never-expose") {
			t.Fatalf("account: %d %s", w.Code, w.Body)
		}
		if id == nil && string(field) != "null" || id != nil && string(field) != "42" {
			t.Fatalf("active value: %s", field)
		}
	}
}
