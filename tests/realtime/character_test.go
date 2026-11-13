package realtime_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"gotest/internal/auth"
	"gotest/internal/middleware"
	"gotest/internal/realtime"
)

type characterAuth struct {
	selected atomic.Int64
	revoked  atomic.Bool
}

func (a *characterAuth) CurrentUser(_ context.Context, token string) (*auth.Identity, error) {
	if a.revoked.Load() {
		return nil, auth.ErrUnauthenticated
	}
	if strings.HasPrefix(token, "b") {
		return &auth.Identity{ID: 8, CharacterID: new(11)}, nil
	}
	selected := int(a.selected.Load())
	return &auth.Identity{ID: 7, CharacterID: &selected}, nil
}
func characterSetup(t *testing.T) (*realtime.Hub, *characterAuth, *httptest.Server) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	a := &characterAuth{}
	a.selected.Store(10)
	h := realtime.NewCharacterHub(ctx, a, "http://localhost:5173")
	s := httptest.NewServer(middleware.NewMiddleware(a).Auth(h.ServeHTTP))
	t.Cleanup(func() { cancel(); s.Close() })
	return h, a, s
}
func characterSocket(t *testing.T, s *httptest.Server, query, token, origin string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return websocket.Dial(ctx, "ws"+strings.TrimPrefix(s.URL, "http")+"/api/battles/events"+query, &websocket.DialOptions{HTTPHeader: http.Header{"Cookie": {"session_token=" + strings.Repeat(token, 64)}, "Origin": {origin}}})
}
func TestCharacterSocketScopeAndSelection(t *testing.T) {
	h, a, s := characterSetup(t)
	for _, tc := range []struct {
		query, origin string
		status        int
	}{
		{"?account_id=7", s.URL, 400},
		{"?account_id=7&character_id=0", s.URL, 400},
		{"?account_id=7&character_id=10&character_id=11", s.URL, 400},
		{"?account_id=7&character_id=11", s.URL, 409},
		{"?account_id=8&character_id=10", s.URL, 409},
		{"?account_id=7&character_id=10", "https://evil.example", 403},
	} {
		c, response, err := characterSocket(t, s, tc.query, "a", tc.origin)
		if c != nil {
			c.CloseNow()
		}
		if err == nil || response == nil || response.StatusCode != tc.status {
			t.Fatalf("%+v: %v %v", tc, response, err)
		}
	}
	c, _, err := characterSocket(t, s, "?account_id=7&character_id=10", "a", s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	read(t, c)
	messages := make(chan error, 4)
	go func() {
		for {
			_, _, err := c.Read(context.Background())
			messages <- err
			if err != nil {
				return
			}
		}
	}()
	h.NotifyCharacter(11)
	select {
	case err := <-messages:
		t.Fatalf("another character notified socket: %v", err)
	case <-time.After(75 * time.Millisecond):
	}
	h.NotifyCharacter(10)
	select {
	case err := <-messages:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("missing own event")
	}
	a.selected.Store(11)
	h.NotifyCharacter(10)
	select {
	case err := <-messages:
		if err == nil {
			t.Fatal("changed selection still receives events")
		}
	case <-time.After(time.Second):
		t.Fatal("stale socket not closed")
	}
}
func TestCharacterSocketRevocation(t *testing.T) {
	h, a, s := characterSetup(t)
	c, _, err := characterSocket(t, s, "?account_id=7&character_id=10", "a", s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	read(t, c)
	a.revoked.Store(true)
	h.NotifyCharacter(10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, _, err = c.Read(ctx); err == nil || ctx.Err() != nil {
		t.Fatalf("revoked socket remained open: %v", err)
	}
}
