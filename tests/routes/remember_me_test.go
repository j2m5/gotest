package routes_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
)

func TestLoginCookiePersistenceIsOptIn(t *testing.T) {
	for _, tc := range []struct {
		name, extra string
		persistent  bool
	}{
		{"omitted", "", false},
		{"false", `,"remember_me":false`, false},
		{"null", `,"remember_me":null`, false},
		{"true", `,"remember_me":true`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := request(router(t, &authStub{}, true, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password"`+tc.extra+`}`, nil)
			cookies := w.Result().Cookies()
			if w.Code != http.StatusOK || len(cookies) != 1 {
				t.Fatalf("login: %d %s", w.Code, w.Body)
			}
			c := cookies[0]
			if c.Name != "session_token" || c.Value == "" || c.Path != "/" || c.Domain != "" || !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
				t.Fatalf("cookie scope/security: %+v", c)
			}
			if tc.persistent {
				if c.MaxAge < 3590 || c.MaxAge > 3600 || c.Expires.IsZero() {
					t.Fatalf("missing persistent lifetime: %+v", c)
				}
			} else if c.MaxAge != 0 || !c.Expires.IsZero() || strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=") {
				t.Fatalf("ordinary login must issue a session cookie: %+v", c)
			}
			assertPublicUserFields(t, w)
		})
	}
}

func TestRememberedLoginFailureDoesNotIssueCookie(t *testing.T) {
	for _, err := range []error{auth.ErrInvalidCredentials, auth.ErrUnverified, errors.New("storage unavailable")} {
		w := request(router(t, &authStub{err: err}, true, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password","remember_me":true}`, nil)
		if w.Code < 400 || len(w.Result().Cookies()) != 0 {
			t.Fatalf("failed login issued a cookie: %d", w.Code)
		}
	}
}

func TestRememberedCookieAlreadyExpiredDuringLoginIsCleared(t *testing.T) {
	w := request(router(t, &authStub{expiresAt: time.Now().Add(-time.Millisecond)}, true, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password","remember_me":true}`, nil)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusOK || len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Fatal("expired persistent session became a browser-session cookie")
	}
}

func TestLoginRejectsInvalidRememberMeTypes(t *testing.T) {
	for _, value := range []string{`"true"`, `1`, `[]`, `{}`} {
		t.Run(value, func(t *testing.T) {
			s := &authStub{}
			w := request(router(t, s, false, 30), "POST", "/api/auth/login", `{"email":"p@example.test","password":"password","remember_me":`+value+`}`, nil)
			if w.Code != http.StatusBadRequest || s.calls != 0 || len(w.Result().Cookies()) != 0 {
				t.Fatalf("invalid flag reached auth: %d, calls=%d", w.Code, s.calls)
			}
		})
	}
}
