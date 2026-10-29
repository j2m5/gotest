package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/notification"
	"gotest/internal/presence"
	"gotest/internal/routes"
)

type flowMail struct {
	once   sync.Once
	tokens chan string
}

func (m *flowMail) verificationTokens() chan string {
	m.once.Do(func() { m.tokens = make(chan string, 16) })
	return m.tokens
}

func (m *flowMail) Deliver(_ context.Context, delivery notification.Delivery) error {
	if delivery.Kind == notification.Verification {
		select {
		case m.verificationTokens() <- delivery.Token:
		default:
		}
	}
	return nil
}

func (m *flowMail) waitVerification(t *testing.T) string {
	t.Helper()
	select {
	case token := <-m.verificationTokens():
		return token
	case <-time.After(5 * time.Second):
		t.Fatal("durable worker did not deliver verification")
		return ""
	}
}

func characterFlowRouter(t *testing.T, s *testStorage, mail notification.Sender, options ...auth.Options) (http.Handler, *auth.Service) {
	t.Helper()
	var authOptions auth.Options
	if len(options) > 0 {
		authOptions = options[0]
	}
	authService := auth.NewService(s.AuthRepository, authOptions)
	worker := notification.NewWorker(db.NewMailQueue(s.pool), testCipher(), mail, func(ctx context.Context, job notification.Job, content notification.Content) error {
		return authService.ProcessPasswordReset(ctx, auth.QueuedResetRequest{ID: job.ID, Lease: job.Lease, Email: content.Email})
	}, nil)
	workerCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); worker.Run(workerCtx) }()
	t.Cleanup(func() { cancel(); <-done })
	characters := character.NewService(db.NewCharacterRepository(s.pool))
	h, err := routes.New(handlers.NewHandler(authService, false, authService), handlers.NewCharacterHandler(characters), middleware.NewMiddleware(authService), routes.Options{StaticDir: filepath.Join("..", "..", "frontend", "dist"), Presence: handlers.NewPresenceHandler(presence.NewService(db.NewPresenceRepository(s.pool)))})
	if err != nil {
		t.Fatal(err)
	}
	return h, authService
}

func flowRequest(t *testing.T, h http.Handler, method, path, body string, cookie *http.Cookie, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: status %d, want %d; body %s", method, path, w.Code, status, w.Body)
	}
	return w
}

func TestCharacterLifecycleThroughHTTP(t *testing.T) {
	s, _ := characterTestStorage(t, true)
	mail := &flowMail{}
	h, _ := characterFlowRouter(t, s, mail)
	flowRequest(t, h, "POST", "/api/auth/register", `{"email":"flow@example.test","login":"flow","password":"browser-password"}`, nil, 201)
	flowRequest(t, h, "POST", "/api/auth/verify", fmt.Sprintf(`{"token":%q}`, mail.waitVerification(t)), nil, 200)
	login := func() *http.Cookie {
		w := flowRequest(t, h, "POST", "/api/auth/login", `{"email":"flow@example.test","password":"browser-password"}`, nil, 200)
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatal("missing login cookie")
		}
		return cookies[0]
	}
	first, second := login(), login()
	roster := func(cookie *http.Cookie) character.Roster {
		w := flowRequest(t, h, "GET", "/api/characters", "", cookie, 200)
		var value character.Roster
		if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	if got := roster(first); got.Characters == nil || len(got.Characters) != 0 || got.ActiveCharacterID != nil {
		t.Fatalf("new account roster: %+v", got)
	}
	w := flowRequest(t, h, "GET", "/api/characters/catalog", "", first, 200)
	var catalog character.Catalog
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Classes) != 8 || len(catalog.Factions) != 2 {
		t.Fatal("incomplete catalog")
	}
	class := catalog.Classes[0]
	create := func(name string, status int) *httptest.ResponseRecorder {
		body, _ := json.Marshal(character.CreateInput{Name: name, FactionID: class.FactionID, ClassID: class.ID})
		return flowRequest(t, h, "POST", "/api/characters", string(body), first, status)
	}
	w = create("  Дарт   Малгус  ", 201)
	var created struct {
		Character character.Character `json:"character"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Character.Name != "Дарт Малгус" || created.Character.Level != 1 {
		t.Fatalf("created: %+v", created.Character)
	}
	create("дарт малгус", 409)
	create("Лорд Реван", 201)
	if got := roster(first); len(got.Characters) != 2 || got.ActiveCharacterID != nil {
		t.Fatalf("creation auto-selected or exposed another account: %+v", got)
	}
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":1}`, first, 404)
	flowRequest(t, h, "PUT", "/api/characters/active", fmt.Sprintf(`{"character_id":%d}`, created.Character.ID), first, 200)
	if got := roster(second); got.ActiveCharacterID == nil || *got.ActiveCharacterID != created.Character.ID {
		t.Fatalf("choice not shared: %+v", got)
	}
	w = flowRequest(t, h, "GET", "/api/auth/me", "", second, 200)
	var me struct {
		User struct {
			CharacterID *int `json:"character_id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil || me.User.CharacterID == nil || *me.User.CharacterID != created.Character.ID {
		t.Fatalf("account selection: %v %s", err, w.Body)
	}
	flowRequest(t, h, "PUT", "/api/characters/active", `{"character_id":null}`, second, 200)
	if got := roster(first); got.ActiveCharacterID != nil {
		t.Fatal("clear was not shared")
	}
	flowRequest(t, h, "PUT", "/api/characters/active", fmt.Sprintf(`{"character_id":%d}`, created.Character.ID), first, 200)
	flowRequest(t, h, "POST", "/api/auth/logout", "", first, 200)
	flowRequest(t, h, "GET", "/api/characters", "", first, 401)
	if got := roster(second); got.ActiveCharacterID != nil || len(got.Characters) != 2 {
		t.Fatalf("logout did not preserve second session and clear choice: %+v", got)
	}
	flowRequest(t, h, "PUT", "/api/characters/active", fmt.Sprintf(`{"character_id":%d}`, created.Character.ID), first, 401)
}
