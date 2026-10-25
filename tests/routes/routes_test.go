package routes_test

import (
	"context"
	"encoding/json"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/config"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/models"
	"gotest/internal/routes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type authStub struct {
	err               error
	calls             int
	revoked, verified string
	user              *models.User
}

func (s *authStub) Register(context.Context, auth.RegisterInput) error { s.calls++; return s.err }
func (s *authStub) Login(context.Context, string, string) (*auth.LoginResult, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &auth.LoginResult{User: &models.User{ID: 1, Email: "p@example.test", Login: "player", Password: "must-not-leak"}, Token: strings.Repeat("a", 64), ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s *authStub) Logout(_ context.Context, token string) error {
	s.calls++
	s.revoked = token
	return s.err
}
func (s *authStub) VerifyEmail(_ context.Context, token string) error {
	s.calls++
	s.verified = token
	return s.err
}
func (s *authStub) ResendVerification(context.Context, string, string) error { s.calls++; return s.err }
func (s *authStub) CurrentUser(context.Context, string) (*models.User, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if s.user == nil {
		return nil, auth.ErrUnauthenticated
	}
	return s.user, nil
}

func router(t *testing.T, s *authStub, secure bool, limit int) http.Handler {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]string{"index.html": "<html>SPA</html>", "assets/app.js": "console.log('app')"} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	h, err := routes.New(handlers.NewHandler(s, secure), middleware.NewMiddleware(s), routes.Options{StaticDir: dir, TrustedOrigin: "http://localhost:5173", AuthLimit: limit})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func request(h http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestVerifyIsPublicAndLogoutRevokesCookie(t *testing.T) {
	s := &authStub{}
	h := router(t, s, false, 30)
	w := request(h, "POST", "/api/auth/verify", `{"token":"abc"}`, nil)
	if w.Code != 200 || s.verified != "abc" || w.Header().Get("Location") != "" {
		t.Fatalf("public verify: %d %s", w.Code, w.Body)
	}
	w = request(h, "POST", "/api/auth/logout", "", &http.Cookie{Name: "session_token", Value: "existing-session"})
	if w.Code != 200 || s.revoked != "existing-session" {
		t.Fatalf("logout: %d %s token=%s", w.Code, w.Body, s.revoked)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 || !cookies[0].HttpOnly {
		t.Fatalf("logout cookie: %#v", cookies)
	}
	w = request(h, "POST", "/api/auth/logout", "", nil)
	if w.Code != 200 {
		t.Fatalf("anonymous logout: %d", w.Code)
	}
}
func TestLogoutDoesNotPretendFailedRevocationSucceeded(t *testing.T) {
	s := &authStub{err: errors.New("private database details")}
	h := router(t, s, false, 30)
	w := request(h, "POST", "/api/auth/logout", "", &http.Cookie{Name: "session_token", Value: "session"})
	if w.Code != 500 || w.Header().Get("Set-Cookie") != "" || strings.Contains(w.Body.String(), "private database") {
		t.Fatalf("failed revoke: %d %s", w.Code, w.Body)
	}
}
func TestLoginCookieAndSafeUserResponse(t *testing.T) {
	w := request(router(t, &authStub{}, true, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password"}`, nil)
	cookies := w.Result().Cookies()
	if w.Code != 200 || len(cookies) != 1 {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	c := cookies[0]
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge < 3590 || c.Expires.IsZero() {
		t.Fatalf("unsafe cookie: %#v", c)
	}
	if strings.Contains(w.Body.String(), "must-not-leak") || strings.Contains(w.Body.String(), "Password") {
		t.Fatal("password leaked")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("auth response must not be cached")
	}
}
func TestServiceErrorsHaveStableJSONStatuses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"credentials", auth.ErrInvalidCredentials, 401}, {"unverified", auth.ErrUnverified, 403}, {"conflict", auth.ErrConflict, 409}, {"token", auth.ErrInvalidToken, 400}, {"verified", auth.ErrAlreadyVerified, 409}, {"validation", &auth.ValidationError{Messages: []string{"invalid email"}}, 422}, {"storage", errors.New("private SQL failure"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := request(router(t, &authStub{err: tc.err}, false, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password"}`, nil)
			var body map[string]any
			if w.Code != tc.status || json.Unmarshal(w.Body.Bytes(), &body) != nil || body["error"] == nil || strings.Contains(w.Body.String(), "private SQL") {
				t.Fatalf("error mapping: %d %s", w.Code, w.Body)
			}
		})
	}
}
func TestDeliveryFailureReportsCreatedAccountAndAllowsResend(t *testing.T) {
	h := router(t, &authStub{err: auth.ErrDelivery}, false, 30)
	w := request(h, "POST", "/api/auth/register", `{"email":"p@example.test","login":"player","password":"password"}`, nil)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"verification_sent":false`) {
		t.Fatalf("registration state lost: %d %s", w.Code, w.Body)
	}
	w = request(h, "POST", "/api/auth/resend-verification", `{"email":"p@example.test","password":"password"}`, nil)
	if w.Code != 503 {
		t.Fatalf("delivery retry: %d", w.Code)
	}
}
func TestMalformedBodiesAreRejectedBeforeService(t *testing.T) {
	for _, tc := range []struct {
		name, body, contentType string
		status                  int
	}{
		{"empty", "", "application/json", 400}, {"invalid", "{", "application/json", 400}, {"null", "null", "application/json", 400}, {"extra", `{} {}`, "application/json", 400}, {"array", `[]`, "application/json", 400}, {"unknown", `{"extra":true}`, "application/json", 400}, {"type", `{}`, "text/plain", 415}, {"large", `{"email":"` + strings.Repeat("x", 17000) + `"}`, "application/json", 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &authStub{}
			h := router(t, s, false, 30)
			r := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status || s.calls != 0 {
				t.Fatalf("bad input: %d %s calls=%d", w.Code, w.Body, s.calls)
			}
		})
	}
}
func TestMeDistinguishesMissingSessionFromDatabaseFailure(t *testing.T) {
	s := &authStub{}
	h := router(t, s, false, 30)
	w := request(h, "GET", "/api/auth/me", "", nil)
	if w.Code != 401 || w.Header().Get("Location") != "" {
		t.Fatalf("unauthorized me: %d %s", w.Code, w.Body)
	}
	s.err = errors.New("private database details")
	w = request(h, "GET", "/api/auth/me", "", &http.Cookie{Name: "session_token", Value: "some-token"})
	if w.Code != 500 || strings.Contains(w.Body.String(), "private database") {
		t.Fatalf("db error: %d %s", w.Code, w.Body)
	}
	s.err = nil
	s.user = &models.User{ID: 2, Login: "player", Email: "p@example.test", Password: "hidden"}
	w = request(h, "GET", "/api/auth/me", "", &http.Cookie{Name: "session_token", Value: "some-token"})
	if w.Code != 200 || strings.Contains(w.Body.String(), "hidden") {
		t.Fatalf("me: %d %s", w.Code, w.Body)
	}
}
func TestRoutingSeparatesAPIAssetsAndSPA(t *testing.T) {
	h := router(t, &authStub{}, false, 30)
	for _, tc := range []struct {
		method, path string
		status       int
		content      string
	}{
		{"GET", "/api/missing", 404, "application/json"}, {"GET", "/api", 404, "application/json"}, {"GET", "/api/auth/login", 405, "application/json"}, {"GET", "/assets/app.js", 200, "javascript"}, {"GET", "/assets/missing.js", 404, "text/plain"}, {"GET", "/verify?token=x", 200, "text/html"},
	} {
		w := request(h, tc.method, tc.path, "", nil)
		if w.Code != tc.status || !strings.Contains(w.Header().Get("Content-Type"), tc.content) {
			t.Errorf("%s: %d %s", tc.path, w.Code, w.Header().Get("Content-Type"))
		}
	}
}
func TestCrossOriginAndRateLimitProtectMutations(t *testing.T) {
	s := &authStub{}
	h := router(t, s, false, 2)
	r := httptest.NewRequest("POST", "http://example.test/api/auth/logout", nil)
	r.Header.Set("Origin", "https://attacker.test")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 || s.calls != 0 || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("cross origin: %d %s", w.Code, w.Body)
	}
	for i := 0; i < 3; i++ {
		w = request(h, "POST", "/api/auth/login", `{"email":"p@example.test","password":"password"}`, nil)
	}
	if w.Code != 429 || w.Header().Get("Retry-After") == "" || s.calls != 2 {
		t.Fatalf("rate limit: %d %s calls=%d", w.Code, w.Body, s.calls)
	}
}

func TestConfiguredPublicURLBuildsRouterAndKeepsCookieSecure(t *testing.T) {
	for _, publicURL := range []string{"https://example.test/", "HTTPS://example.test"} {
		t.Run(publicURL, func(t *testing.T) {
			for _, key := range []string{"APP_PORT", "SESSION_TTL", "VERIFICATION_TTL", "SMTP_TIMEOUT"} {
				t.Setenv(key, "")
			}
			t.Setenv("DATABASE_URL", "postgres://unused/test")
			t.Setenv("APP_ENV", "production")
			t.Setenv("PUBLIC_URL", publicURL)
			cfg, err := config.Load()
			if err != nil {
				t.Fatal(err)
			}
			s := &authStub{}
			h, err := routes.New(handlers.NewHandler(s, strings.HasPrefix(cfg.PublicURL, "https://")), middleware.NewMiddleware(s), routes.Options{StaticDir: t.TempDir(), TrustedOrigin: cfg.PublicURL})
			if err != nil {
				t.Fatalf("accepted config cannot start router: %v", err)
			}
			r := httptest.NewRequest("POST", "http://internal/api/auth/login", strings.NewReader(`{"email":"p@example.test","password":"password"}`))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Origin", "https://example.test")
			r.Header.Set("Sec-Fetch-Site", "cross-site")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 200 {
				t.Fatalf("configured origin rejected: %d %s", w.Code, w.Body)
			}
			if cookies := w.Result().Cookies(); len(cookies) != 1 || !cookies[0].Secure {
				t.Fatalf("HTTPS configuration issued insecure cookie: %#v", cookies)
			}
		})
	}
}
