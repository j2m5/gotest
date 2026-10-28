package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest/internal/auth"
)

func accountRequest(h http.Handler, method, path, body string, cookie *http.Cookie, values ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	for _, value := range values {
		r.Header.Add("X-Account-ID", value)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestAccountPreconditionRejectsChangedAccountBeforeCharacterWork(t *testing.T) {
	for _, endpoint := range []struct{ method, path, body string }{
		{"GET", "/api/characters", ""},
		{"GET", "/api/characters/catalog", ""},
		{"POST", "/api/characters", `{"name":"Malgus","faction_id":2,"class_id":8}`},
		{"PUT", "/api/characters/active", `{"character_id":null}`},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			s := &characterStub{}
			w := accountRequest(characterRouter(t, s), endpoint.method, endpoint.path, endpoint.body, characterCookie(), "6")
			if w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"account_changed"`) || s.calls != 0 || len(w.Result().Cookies()) != 0 {
				t.Fatalf("stale account performed work: status=%d calls=%d body=%s", w.Code, s.calls, w.Body)
			}
		})
	}
}

func TestAccountPreconditionValidationAndCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []string
		status int
	}{
		{"absent", nil, 200}, {"matching", []string{"7"}, 200},
		{"empty", []string{""}, 400}, {"zero", []string{"0"}, 400},
		{"negative", []string{"-7"}, 400}, {"signed", []string{"+7"}, 400},
		{"overflow", []string{"2147483648"}, 400}, {"fraction", []string{"7.0"}, 400},
		{"combined", []string{"7,7"}, 400}, {"duplicate", []string{"7", "7"}, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &characterStub{}
			w := accountRequest(characterRouter(t, s), "GET", "/api/characters", "", characterCookie(), tc.values...)
			if w.Code != tc.status || (tc.status != 200 && (s.calls != 0 || !strings.Contains(w.Body.String(), `"code":"invalid_request"`))) {
				t.Fatalf("precondition: %d calls=%d %s", w.Code, s.calls, w.Body)
			}
		})
	}
}

func TestAccountPreconditionDoesNotLogOutReplacementAccount(t *testing.T) {
	s := &authStub{user: &auth.Identity{ID: 2}}
	h := router(t, s, false, 1000)
	w := accountRequest(h, "POST", "/api/auth/logout", "", characterCookie(), "1")
	if w.Code != 409 || s.revoked != "" || len(w.Result().Cookies()) != 0 {
		t.Fatalf("replacement session was logged out: %d revoked=%t cookies=%d", w.Code, s.revoked != "", len(w.Result().Cookies()))
	}
	w = accountRequest(h, "POST", "/api/auth/logout", "", characterCookie(), "2")
	if w.Code != 200 || s.revoked != characterCookie().Value || len(w.Result().Cookies()) != 1 {
		t.Fatalf("matching logout failed: %d", w.Code)
	}
}

func TestAccountPreconditionLogoutStillClearsAbsentOrExpiredSession(t *testing.T) {
	for _, cookie := range []*http.Cookie{nil, characterCookie()} {
		s := &authStub{}
		w := accountRequest(router(t, s, false, 1000), "POST", "/api/auth/logout", "", cookie, "1")
		if w.Code != 200 || len(w.Result().Cookies()) != 1 || w.Result().Cookies()[0].MaxAge != -1 {
			t.Fatalf("idempotent logout changed: %d %s", w.Code, w.Body)
		}
	}
}
