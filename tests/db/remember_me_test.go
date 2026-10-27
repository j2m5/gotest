package db_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"gotest/internal/auth"
)

func TestLoginSessionModesThroughHTTPAndPostgres(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mail := &flowMail{}
	h, _ := characterFlowRouter(t, s, mail, auth.Options{SessionTTL: 2 * time.Hour, RememberMeTTL: 48 * time.Hour})
	flowRequest(t, h, "POST", "/api/auth/register", `{"email":"remember@example.test","login":"remember","password":"browser-password"}`, nil, 201)
	flowRequest(t, h, "POST", "/api/auth/verify", fmt.Sprintf(`{"token":%q}`, mail.token), nil, 200)
	login := func(remember bool, ttl time.Duration) *http.Cookie {
		t.Helper()
		before := time.Now()
		w := flowRequest(t, h, "POST", "/api/auth/login", fmt.Sprintf(`{"email":"remember@example.test","password":"browser-password","remember_me":%t}`, remember), nil, 200)
		if len(w.Result().Cookies()) != 1 {
			t.Fatal("missing session cookie")
		}
		cookie := w.Result().Cookies()[0]
		var expires time.Time
		var rawCount int
		if err := s.pool.QueryRow(ctx, `SELECT expires_at FROM sessions WHERE token=$1`, auth.TokenDigest(cookie.Value)).Scan(&expires); err != nil {
			t.Fatal(err)
		}
		if expires.Before(before.Add(ttl)) || expires.After(time.Now().Add(ttl)) {
			t.Fatalf("wrong persisted lifetime: %v", expires.Sub(before))
		}
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE token=$1`, cookie.Value).Scan(&rawCount); err != nil || rawCount != 0 {
			t.Fatal("raw session token stored")
		}
		if remember {
			if cookie.MaxAge < int(ttl.Seconds())-5 || !cookie.Expires.Equal(expires.UTC().Truncate(time.Second)) {
				t.Fatal("persistent cookie does not match stored expiry")
			}
		} else if cookie.MaxAge != 0 || !cookie.Expires.IsZero() {
			t.Fatal("ordinary login issued a persistent cookie")
		}
		flowRequest(t, h, "GET", "/api/auth/me", "", cookie, 200)
		var after time.Time
		if err := s.pool.QueryRow(ctx, `SELECT expires_at FROM sessions WHERE token=$1`, auth.TokenDigest(cookie.Value)).Scan(&after); err != nil || !after.Equal(expires) {
			t.Fatal("reading session changed its deadline")
		}
		return cookie
	}
	ordinary, remembered := login(false, 2*time.Hour), login(true, 48*time.Hour)
	for _, cookie := range []*http.Cookie{ordinary, remembered} {
		w := flowRequest(t, h, "POST", "/api/auth/logout", "", cookie, 200)
		cleared := w.Result().Cookies()
		if len(cleared) != 1 || cleared[0].MaxAge != -1 || cleared[0].Name != cookie.Name || cleared[0].Path != cookie.Path {
			t.Fatal("logout failed to clear the session cookie")
		}
		flowRequest(t, h, "GET", "/api/auth/me", "", cookie, 401)
		if cookie == ordinary {
			flowRequest(t, h, "GET", "/api/auth/me", "", remembered, 200)
		}
	}
	for _, remember := range []bool{false, true} {
		ttl := 2 * time.Hour
		if remember {
			ttl = 48 * time.Hour
		}
		cookie := login(remember, ttl)
		if _, err := s.pool.Exec(ctx, `UPDATE sessions SET expires_at=now() - interval '1 second' WHERE token=$1`, auth.TokenDigest(cookie.Value)); err != nil {
			t.Fatal(err)
		}
		flowRequest(t, h, "GET", "/api/auth/me", "", cookie, 401)
	}
}
