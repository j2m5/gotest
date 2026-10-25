package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/clientip"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
)

func proxyLogin(h http.Handler, peer string, forwarded ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"p@example.test","password":"password"}`))
	r.RemoteAddr = peer
	r.Header.Set("Content-Type", "application/json")
	for _, value := range forwarded {
		r.Header.Add("X-Forwarded-For", value)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func proxyRouter(t *testing.T, trusted string, s *authStub) http.Handler {
	t.Helper()
	policy, err := clientip.Parse(trusted)
	if err != nil {
		t.Fatal(err)
	}
	h, err := routes.New(handlers.NewHandler(s, true), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(s), routes.Options{
		StaticDir: t.TempDir(), TrustedOrigin: "https://example.test", AuthLimit: 1, TrustedProxies: policy,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestRateLimitSeparatesClientsBehindTrustedProxy(t *testing.T) {
	h := proxyRouter(t, "10.0.0.0/8", &authStub{err: auth.ErrInvalidCredentials})
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1000", "198.51.100.1"), 401, "invalid_credentials")
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1001", "198.51.100.2"), 401, "invalid_credentials")
	w := proxyLogin(h, "10.0.0.3:1002", "198.51.100.1, 10.0.0.2")
	requireAPIError(t, w, 429, "rate_limited")
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("rate limit has no retry delay")
	}
}

func TestRateLimitCannotBeEvadedWithForgedForwardedPrefix(t *testing.T) {
	h := proxyRouter(t, "10.0.0.0/8", &authStub{err: auth.ErrInvalidCredentials})
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1000", "198.51.100.1, 192.0.2.5, 10.1.1.1"), 401, "invalid_credentials")
	for _, prefix := range []string{"198.51.100.2", "unknown,,", ""} {
		requireAPIError(t, proxyLogin(h, "10.0.0.2:1001", prefix, "192.0.2.5, 10.1.1.1"), 429, "rate_limited")
	}
}

func TestRateLimitIgnoresHeadersFromUntrustedPeers(t *testing.T) {
	for _, trusted := range []string{"", "10.0.0.0/8"} {
		t.Run(trusted, func(t *testing.T) {
			h := proxyRouter(t, trusted, &authStub{err: auth.ErrInvalidCredentials})
			requireAPIError(t, proxyLogin(h, "192.0.2.5:1000", "198.51.100.1"), 401, "invalid_credentials")
			r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
			r.RemoteAddr = "192.0.2.5:2000"
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-Forwarded-For", "198.51.100.2, 10.1.1.1")
			r.Header.Set("X-Real-IP", "198.51.100.3")
			r.Header.Set("Forwarded", "for=198.51.100.4")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			requireAPIError(t, w, 429, "rate_limited")
		})
	}
}

func TestRateLimitCanonicalizesForwardedAddresses(t *testing.T) {
	for _, pair := range [][2]string{
		{"2001:db8::5", "2001:0db8:0:0:0:0:0:5"},
		{"192.0.2.5", "::ffff:192.0.2.5"},
	} {
		h := proxyRouter(t, "10.0.0.0/8", &authStub{err: auth.ErrInvalidCredentials})
		requireAPIError(t, proxyLogin(h, "10.0.0.2:1000", pair[0]), 401, "invalid_credentials")
		requireAPIError(t, proxyLogin(h, "10.0.0.2:1001", pair[1]), 429, "rate_limited")
	}
}

func TestRateLimitSharesPeerBucketForMalformedChains(t *testing.T) {
	h := proxyRouter(t, "10.0.0.0/8", &authStub{err: auth.ErrInvalidCredentials})
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1000", "bad-one"), 401, "invalid_credentials")
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1001", "bad-two"), 429, "rate_limited")
	requireAPIError(t, proxyLogin(h, "10.0.0.2:1002"), 429, "rate_limited")
}

func TestProxyHeadersDoNotChangeOriginOrCookieSecurity(t *testing.T) {
	h := proxyRouter(t, "10.0.0.0/8", &authStub{})
	for _, origin := range []string{"https://attacker.test", "https://example.test"} {
		r := httptest.NewRequest("POST", "http://internal/api/auth/login", strings.NewReader(`{"email":"p@example.test","password":"password"}`))
		r.RemoteAddr = "10.0.0.2:1000"
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("Sec-Fetch-Site", "cross-site")
		r.Header.Set("X-Forwarded-For", "198.51.100.1")
		r.Header.Set("X-Forwarded-Host", "attacker.test")
		r.Header.Set("X-Forwarded-Proto", "http")
		r.Header.Set("Forwarded", "for=198.51.100.1;host=attacker.test;proto=http")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if origin == "https://attacker.test" {
			requireAPIError(t, w, 403, "origin_forbidden")
			continue
		}
		if w.Code != 200 {
			t.Fatalf("configured origin rejected: %d %s", w.Code, w.Body)
		}
		if cookies := w.Result().Cookies(); len(cookies) != 1 || !cookies[0].Secure {
			t.Fatalf("forwarded proto changed configured cookie security: %#v", cookies)
		}
		if r.Host != "internal" || r.URL.Scheme != "http" || r.RemoteAddr != "10.0.0.2:1000" {
			t.Fatal("proxy headers rewrote transport or origin information")
		}
	}
}

func TestRateLimitCanonicalizesPeerAddresses(t *testing.T) {
	for _, tc := range []struct{ name, first, second string }{
		{"IPv6", "[2001:db8::5]:1000", "[2001:0db8:0:0:0:0:0:5]:2000"},
		{"mapped IPv4", "192.0.2.5:1000", "[::ffff:192.0.2.5]:2000"},
		{"invalid peers", "invalid-one", "invalid-two"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := router(t, &authStub{err: auth.ErrInvalidCredentials}, false, 1)
			requireAPIError(t, proxyLogin(h, tc.first), 401, "invalid_credentials")
			requireAPIError(t, proxyLogin(h, tc.second), 429, "rate_limited")
		})
	}
}
