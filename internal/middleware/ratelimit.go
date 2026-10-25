package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateWindow struct {
	count int
	until time.Time
}
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateWindow
	limit   int
	period  time.Duration
	now     func() time.Time
}

func NewRateLimiter(limit int, period time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 30
	}
	return &RateLimiter{entries: make(map[string]rateWindow), limit: limit, period: period, now: time.Now}
}
func (l *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		now := l.now()
		l.mu.Lock()
		entry, found := l.entries[ip]
		if !found || !entry.until.After(now) {
			// Bound memory even when clients continually vary their address.
			if len(l.entries) >= 10000 {
				for key, value := range l.entries {
					if !value.until.After(now) {
						delete(l.entries, key)
					}
				}
			}
			if !found && len(l.entries) >= 10000 {
				l.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				writeError(w, 429, "Слишком много запросов. Повторите позже")
				return
			}
			entry = rateWindow{until: now.Add(l.period)}
		}
		entry.count++
		l.entries[ip] = entry
		l.mu.Unlock()
		if entry.count > l.limit {
			seconds := int(entry.until.Sub(now).Seconds()) + 1
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			writeError(w, 429, "Слишком много запросов. Повторите позже")
			return
		}
		next.ServeHTTP(w, r)
	})
}
