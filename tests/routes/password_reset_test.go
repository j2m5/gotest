package routes_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/validation"
)

func TestPasswordResetRoutesArePublicAndRateLimited(t *testing.T) {
	for _, path := range []string{"/api/auth/forgot-password", "/api/auth/reset-password"} {
		t.Run(path, func(t *testing.T) {
			s := &authStub{}
			h := router(t, s, true, 1)
			body := `{"email":"player@example.test"}`
			if strings.HasSuffix(path, "/reset-password") {
				body = `{"token":"token","password":"new-password"}`
			}
			w := request(h, "POST", path, body, nil)
			if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("public reset route: %d %s", w.Code, w.Body)
			}
			if w := request(h, "POST", path, body, nil); w.Code != 429 {
				t.Fatalf("reset route not rate limited: %d", w.Code)
			}
			if w := request(router(t, &authStub{}, false, 30), "GET", path, "", nil); w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("GET consumed reset: %d", w.Code)
			}
		})
	}
}

func TestResetPasswordFailuresAndNotificationOutcome(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"token", auth.ErrInvalidResetToken, 400, "password_reset_token_invalid"},
		{"password", &validation.Error{Fields: validation.Fields{"password": {"invalid"}}}, 422, "validation_failed"},
		{"storage", errors.New("private database detail"), 500, "internal_error"},
		{"notice", errors.Join(auth.ErrPasswordChangedDelivery, errors.New("private SMTP detail")), 200, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := request(router(t, &authStub{err: tc.err}, true, 30), "POST", "/api/auth/reset-password", `{"token":"token","password":"password"}`, nil)
			if w.Code != tc.status || strings.Contains(w.Body.String(), "private") || !strings.Contains(w.Body.String(), tc.code) {
				t.Fatalf("reset response: %d %s", w.Code, w.Body)
			}
			if tc.status == 200 {
				cookies := w.Result().Cookies()
				if len(cookies) != 1 || cookies[0].MaxAge != -1 || !cookies[0].HttpOnly || !cookies[0].Secure {
					t.Fatal("success did not clear cookie")
				}
			} else if len(w.Result().Cookies()) != 0 {
				t.Fatal("failed reset cleared current session")
			}
		})
	}
	for _, err := range []error{auth.ErrResetUnavailable, errors.New("private queue details")} {
		w := request(router(t, &authStub{err: err}, false, 30), "POST", "/api/auth/forgot-password", `{"email":"person@example.test"}`, nil)
		if w.Code != 503 || strings.Contains(w.Body.String(), "private") || len(w.Result().Cookies()) != 0 {
			t.Fatal("unavailable queue contract")
		}
	}
}
