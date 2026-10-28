package routes_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/routers/legacy"

	"gotest/internal/auth"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
	"gotest/internal/social"
)

type socialStub struct {
	err           error
	actor         social.Actor
	sent          int
	notifications int
	filled        bool
}

func (s *socialStub) Moderators(_ context.Context, a social.Actor) (*social.ModeratorList, error) {
	s.actor = a
	list := &social.ModeratorList{Moderators: []social.Moderator{}}
	if s.filled {
		list.Moderators = []social.Moderator{{UserID: 2, Login: "admin", Role: "administrator"}, {UserID: 3, Login: "mod", Role: "moderator", CharacterName: new("Watcher")}}
	}
	return list, s.err
}

func (s *socialStub) Chat(_ context.Context, a social.Actor, _ int64) (*social.ChatPage, error) {
	s.actor = a
	page := &social.ChatPage{Messages: []social.Message{}}
	if s.filled {
		body := "Hello"
		now := time.Now()
		page.Messages = []social.Message{{ID: 2, UserID: 2, CharacterID: 3, AuthorName: "Player", AuthorRole: "user", Body: &body, CreatedAt: now}, {ID: 1, UserID: 2, CharacterID: 3, AuthorName: "Player", AuthorRole: "user", CreatedAt: now, DeletedAt: &now}}
		page.ChatBan = s.sampleBan()
		page.NextBeforeID = new(int64(1))
	}
	return page, s.err
}

func (s *socialStub) sampleBan() *social.Ban {
	now := time.Now()
	expires := now.Add(time.Hour)
	return &social.Ban{ID: 1, UserID: 2, Scope: "chat", Reason: "Spam", CreatedByUserID: 1, CreatedAt: now, ExpiresAt: &expires}
}
func (s *socialStub) Send(_ context.Context, a social.Actor, in social.SendInput) (*social.Message, error) {
	s.actor = a
	s.sent++
	return &social.Message{ID: 1, UserID: a.UserID, CharacterID: in.CharacterID, AuthorName: "Player", AuthorRole: "user", Body: &in.Body, CreatedAt: time.Now()}, s.err
}
func (s *socialStub) Delete(_ context.Context, a social.Actor, _ social.DeleteInput) error {
	s.actor = a
	return s.err
}
func (s *socialStub) Bans(_ context.Context, a social.Actor, _ int64, _ int) (*social.BanPage, error) {
	s.actor = a
	p := &social.BanPage{Bans: []social.Ban{}}
	if s.filled {
		p.Bans = append(p.Bans, *s.sampleBan())
	}
	return p, s.err
}
func (s *socialStub) Ban(_ context.Context, a social.Actor, in social.BanInput) (*social.Ban, error) {
	s.actor = a
	return &social.Ban{ID: 1, UserID: in.UserID, Scope: in.Scope, Reason: in.Reason, CreatedByUserID: a.UserID, CreatedAt: time.Now()}, s.err
}
func (s *socialStub) Revoke(_ context.Context, a social.Actor, _ social.RevokeInput) error {
	s.actor = a
	return s.err
}
func (s *socialStub) Actions(_ context.Context, a social.Actor, _ int64) (*social.ActionPage, error) {
	s.actor = a
	p := &social.ActionPage{Actions: []social.Action{}}
	if s.filled {
		p.Actions = append(p.Actions, social.Action{ID: 1, ActorUserID: new(1), UserID: 2, Scope: "chat", Action: "delete_message", Reason: "Spam", MessageID: new(int64(1)), CreatedAt: time.Now()})
	}
	return p, s.err
}

func TestSocialRoutesBindSessionAndNotifyOnlySuccess(t *testing.T) {
	contract, err := legacy.NewRouter(loadOpenAPI(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, body    string
		err           error
		status        int
		notifications int
	}{
		{"/api/chat/messages", `{"body":"Hello","character_id":3}`, nil, 201, 1},
		{"/api/chat/messages", `{"body":"Hello","character_id":3}`, social.ErrChatBanned, 403, 0},
		{"/api/chat/messages", `{"body":"Hello","character_id":3}`, social.ErrTooFast, 429, 0},
		{"/api/chat/messages", `{"body":"Hello","character_id":3}`, social.ErrCharacterChanged, 409, 0},
		{"/api/chat/delete", `{"message_id":1,"reason":"spam"}`, social.ErrForbidden, 403, 0},
		{"/api/chat/delete", `{"message_id":1,"reason":"spam"}`, nil, 200, 1},
		{"/api/moderation/bans", `{"user_id":3,"scope":"account","reason":"spam"}`, nil, 400, 0},
		{"/api/moderation/bans", `{"user_id":3,"scope":"account","duration_hours":null,"reason":"spam"}`, nil, 201, 1},
		{"/api/moderation/revoke", `{"ban_id":1,"reason":"appeal"}`, nil, 200, 1},
	} {
		t.Run(tc.path+tc.body+string(rune(tc.status)), func(t *testing.T) {
			service := &socialStub{err: tc.err}
			a := &authStub{user: &auth.Identity{ID: 7}}
			h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(a), routes.Options{
				StaticDir: t.TempDir(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Social: handlers.NewSocialHandler(service, func() { service.notifications++ }),
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			r.AddCookie(characterCookie())
			r.Header.Set("X-Account-ID", "7")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if err := validateContractResponse(contractInput(t, contract, r), w.Code, w.Header(), w.Body.Bytes()); err != nil {
				t.Fatal(err)
			}
			if service.notifications != tc.notifications {
				t.Fatalf("notifications=%d", service.notifications)
			}
			if tc.status != 400 && (service.actor.UserID != 7 || service.actor.SessionHash == "") {
				t.Fatalf("actor=%+v", service.actor)
			}
		})
	}
}

func TestSocialQueryRejectsAmbiguousAndOverflowIDs(t *testing.T) {
	service := &socialStub{}
	a := &authStub{user: &auth.Identity{ID: 7}}
	h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(a), routes.Options{StaticDir: t.TempDir(), Social: handlers.NewSocialHandler(service, nil)})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/chat/messages?before_id=-1", "/api/chat/messages?before_id=", "/api/chat/messages?before_id=1&before_id=2", "/api/chat/messages?before_id=9223372036854775808", "/api/moderation/bans?user_id=2147483648"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(characterCookie())
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 400 {
			t.Errorf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
}

func TestModeratorDirectoryRequiresMatchingAccount(t *testing.T) {
	service := &socialStub{filled: true}
	a := &authStub{user: &auth.Identity{ID: 7, RoleID: auth.RoleUser}}
	h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(a), routes.Options{StaticDir: t.TempDir(), Social: handlers.NewSocialHandler(service, nil)})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		cookie  bool
		account string
		status  int
	}{{false, "", 401}, {true, "8", 409}, {true, "7", 200}} {
		r := httptest.NewRequest("GET", "/api/chat/moderators", nil)
		if tc.cookie {
			r.AddCookie(characterCookie())
		}
		if tc.account != "" {
			r.Header.Set("X-Account-ID", tc.account)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("directory %+v: %d %s", tc, w.Code, w.Body.String())
		}
	}
}
