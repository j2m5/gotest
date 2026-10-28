package realtime_test

import (
	"context"
	"io"
	"log/slog"
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

type authenticator struct{ revoked atomic.Bool }

func (a *authenticator) CurrentUser(context.Context, string) (*auth.Identity, error) {
	if a.revoked.Load() {
		return nil, auth.ErrUnauthenticated
	}
	return &auth.Identity{ID: 7}, nil
}

func setup(t *testing.T) (*realtime.Hub, *authenticator, *httptest.Server, context.CancelFunc) {
	return setupTimeout(t, 0)
}

func setupTimeout(t *testing.T, timeout time.Duration) (*realtime.Hub, *authenticator, *httptest.Server, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	a := &authenticator{}
	h := realtime.New(ctx, a, "http://localhost:5173")
	s := httptest.NewUnstartedServer(middleware.Observe(slog.New(slog.NewTextHandler(io.Discard, nil)), middleware.NewMiddleware(a).Auth(h.ServeHTTP)))
	s.Config.ReadTimeout = timeout
	s.Config.WriteTimeout = timeout
	s.Start()
	t.Cleanup(func() { cancel(); s.Close() })
	return h, a, s, cancel
}

func connect(t *testing.T, s *httptest.Server, query, origin string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := http.Header{"Cookie": {"session_token=" + strings.Repeat("a", 64)}}
	if origin != "" {
		header.Set("Origin", origin)
	}
	return websocket.Dial(ctx, "ws"+strings.TrimPrefix(s.URL, "http")+"/api/chat/events"+query, &websocket.DialOptions{HTTPHeader: header})
}

func read(t *testing.T, c *websocket.Conn) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestSocketRejectsOriginAndStaleAccount(t *testing.T) {
	_, _, s, _ := setup(t)
	for _, tc := range []struct {
		query, origin string
		status        int
	}{
		{"?account_id=7", "https://evil.example", 403},
		{"?account_id=7", "", 403},
		{"?account_id=8", "http://localhost:5173", 409},
		{"", "http://localhost:5173", 400},
		{"?account_id=7&account_id=8", "http://localhost:5173", 400},
	} {
		c, response, err := connect(t, s, tc.query, tc.origin)
		if c != nil {
			c.CloseNow()
		}
		if err == nil || response == nil || response.StatusCode != tc.status {
			t.Fatalf("%+v: response=%v err=%v", tc, response, err)
		}
	}
}

func TestSocketReceivesInvalidationAndClosesOnShutdown(t *testing.T) {
	h, _, s, cancel := setup(t)
	c, _, err := connect(t, s, "?account_id=7", "http://localhost:5173")
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	if got := string(read(t, c)); got != `{"type":"changed"}` {
		t.Fatalf("initial event: %q", got)
	}
	h.Notify()
	if got := string(read(t, c)); got != `{"type":"changed"}` {
		t.Fatalf("invalidation: %q", got)
	}
	cancel()
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if _, _, err := c.Read(ctx); err == nil || ctx.Err() != nil {
		t.Fatalf("shutdown did not promptly close: %v", err)
	}
}

func TestSocketRevocationCheckedBeforeNotification(t *testing.T) {
	h, a, s, _ := setup(t)
	c, _, err := connect(t, s, "?account_id=7", s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	read(t, c)
	a.revoked.Store(true)
	h.Notify()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, _, err := c.Read(ctx); err == nil || ctx.Err() != nil {
		t.Fatalf("revoked socket remained usable: %v", err)
	}
}

func TestSocketConnectionLimitAndRelease(t *testing.T) {
	_, _, s, _ := setup(t)
	var connections []*websocket.Conn
	defer func() {
		for _, c := range connections {
			c.CloseNow()
		}
	}()
	for range 5 {
		c, _, err := connect(t, s, "?account_id=7", s.URL)
		if err != nil {
			t.Fatal(err)
		}
		read(t, c)
		connections = append(connections, c)
	}
	c, response, err := connect(t, s, "?account_id=7", s.URL)
	if c != nil {
		c.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != 429 {
		t.Fatalf("connection limit not enforced: %v %v", response, err)
	}
	connections[0].Close(websocket.StatusNormalClosure, "")
	c, _, err = connect(t, s, "?account_id=7", s.URL)
	if err != nil {
		t.Fatal("connection slot not released:", err)
	}
	connections = append(connections, c)
}

func TestSocketOutlivesHTTPDeadlines(t *testing.T) {
	h, _, s, _ := setupTimeout(t, 500*time.Millisecond)
	c, _, err := connect(t, s, "?account_id=7", s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	read(t, c)
	time.Sleep(750 * time.Millisecond)
	h.Notify()
	read(t, c)
}
