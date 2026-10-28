// Package realtime delivers bounded invalidations for the single-process app.
// Messages and moderation data are fetched through the authenticated HTTP API.
package realtime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
)

type Hub struct {
	lifecycle   context.Context
	auth        middleware.Authenticator
	origin      string
	mu          sync.Mutex
	subscribers map[chan struct{}]int
}

func New(lifecycle context.Context, authenticator middleware.Authenticator, origin string) *Hub {
	if lifecycle == nil {
		lifecycle = context.Background()
	}
	return &Hub{lifecycle: lifecycle, auth: authenticator, origin: origin, subscribers: make(map[chan struct{}]int)}
}

// Notify coalesces bursts. Slow clients cannot block a committed write or grow
// an unbounded event queue. Reconnecting clients always receive a fresh event.
func (h *Hub) Notify() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *Hub) subscribe(userID int) (chan struct{}, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for _, id := range h.subscribers {
		if id == userID {
			count++
		}
	}
	if h.lifecycle.Err() != nil || count >= 5 || len(h.subscribers) >= 500 {
		return nil, false
	}
	ch := make(chan struct{}, 1)
	h.subscribers[ch] = userID
	ch <- struct{}{}
	return ch, true
}

func (h *Hub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subscribers, ch)
}

func (h *Hub) allowedOrigin(r *http.Request) bool {
	values := r.Header.Values("Origin")
	if len(values) != 1 {
		return false
	}
	origin := values[0]
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return false
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return origin == scheme+"://"+r.Host || (h.origin != "" && origin == h.origin)
}

// ServeHTTP is mounted behind Auth. It separately validates Origin because a
// WebSocket upgrade is a GET and therefore not rejected by ordinary CSRF guards.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.allowedOrigin(r) {
		httpapi.WriteError(w, 403, httpapi.CodeOriginForbidden, "Соединение с другого сайта запрещено", nil)
		return
	}
	values := r.URL.Query()["account_id"]
	var expected int64
	if len(values) == 1 {
		expected, _ = strconv.ParseInt(values[0], 10, 32)
	}
	if expected <= 0 {
		httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Укажите аккаунт для соединения", nil)
		return
	}
	user := middleware.UserFromContext(r.Context())
	cookie, err := r.Cookie(middleware.SessionCookie)
	if user == nil || err != nil {
		httpapi.WriteError(w, 401, httpapi.CodeUnauthenticated, "Необходимо войти в аккаунт", nil)
		return
	}
	if int(expected) != user.ID {
		httpapi.WriteError(w, 409, httpapi.CodeAccountChanged, "Аккаунт изменился. Обновите страницу", nil)
		return
	}
	ch, ok := h.subscribe(user.ID)
	if !ok {
		w.Header().Set("Retry-After", "5")
		httpapi.WriteError(w, 429, "rate_limited", "Слишком много открытых соединений чата", nil)
		return
	}
	defer h.unsubscribe(ch)
	// The exact scheme and host have already been checked above. Keep the
	// library's origin check enabled as an additional handshake boundary.
	options := &websocket.AcceptOptions{}
	if h.origin != "" {
		options.OriginPatterns = []string{h.origin}
	}
	c, err := websocket.Accept(w, r, options)
	if err != nil {
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(128)
	// A hijacked connection has its own lifetime; request timeout/cancellation
	// must not tear it down while the application is still running.
	ctx := c.CloseRead(h.lifecycle)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	valid := func() bool {
		check, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		current, err := h.auth.CurrentUser(check, cookie.Value)
		return err == nil && current != nil && current.ID == user.ID
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ch:
			if !valid() {
				return
			}
			write, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Write(write, websocket.MessageText, []byte(`{"type":"changed"}`))
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			if !valid() {
				return
			}
			ping, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Ping(ping)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
