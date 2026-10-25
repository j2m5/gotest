package routes_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/validation"
)

type apiFailure struct {
	Error  string              `json:"error"`
	Code   string              `json:"code"`
	Fields map[string][]string `json:"fields"`
	Errors []string            `json:"errors"`
}

func requireAPIError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) apiFailure {
	t.Helper()
	var body apiFailure
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid error JSON: %v; %s", err, w.Body)
	}
	if w.Code != status || body.Code != code || body.Error == "" {
		t.Fatalf("error contract: status=%d body=%s; want %d/%s", w.Code, w.Body, status, code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("error headers: %v", w.Header())
	}
	if strings.Contains(w.Body.String(), "private-sql-detail") {
		t.Fatal("internal details exposed")
	}
	return body
}

func TestAPIErrorCodesForAuthAndWrappedFailures(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{auth.ErrInvalidCredentials, 401, "invalid_credentials"}, {auth.ErrUnauthenticated, 401, "unauthenticated"},
		{auth.ErrUnverified, 403, "email_unverified"}, {auth.ErrConflict, 409, "account_conflict"},
		{auth.ErrAlreadyVerified, 409, "email_already_verified"}, {auth.ErrInvalidToken, 400, "verification_token_invalid"},
		{auth.ErrDelivery, 503, "mail_delivery_failed"}, {errors.New("private-sql-detail"), 500, "internal_error"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			h := router(t, &authStub{err: fmt.Errorf("operation: %w", tc.err)}, false, 30)
			w := request(h, "POST", "/api/auth/login", `{"email":"a@example.test","password":"password"}`, nil)
			requireAPIError(t, w, tc.status, tc.code)
		})
	}
}

func TestAPIErrorCodesForCharacterFailures(t *testing.T) {
	for _, tc := range []struct {
		err         error
		status      int
		code, field string
	}{
		{character.ErrUnauthenticated, 401, "unauthenticated", ""}, {character.ErrNotFound, 404, "character_not_found", ""},
		{character.ErrNameTaken, 409, "character_name_taken", "name"}, {character.ErrInvalidClass, 422, "character_class_mismatch", "class_id"},
		{errors.New("private-sql-detail"), 500, "internal_error", ""},
	} {
		t.Run(tc.code, func(t *testing.T) {
			w := request(characterRouter(t, &characterStub{err: fmt.Errorf("operation: %w", tc.err)}), "PUT", "/api/characters/active", `{"character_id":1}`, characterCookie())
			body := requireAPIError(t, w, tc.status, tc.code)
			if tc.field != "" && len(body.Fields[tc.field]) == 0 {
				t.Fatalf("missing field error %s: %s", tc.field, w.Body)
			}
		})
	}
}

func TestWrappedValidationPreservesFieldsAndLegacyDetails(t *testing.T) {
	fields := validation.Fields{
		"password": {"Password is too short", "Password contains invalid characters"},
		"email":    {"Email is invalid"},
	}
	err := fmt.Errorf("register: %w", &validation.Error{Fields: fields})
	w := request(router(t, &authStub{err: err}, false, 30), "POST", "/api/auth/register", `{}`, nil)
	body := requireAPIError(t, w, 422, "validation_failed")
	if !reflect.DeepEqual(body.Fields, map[string][]string(fields)) {
		t.Fatalf("field errors changed: %+v", body.Fields)
	}
	wantDetails := []string{"Email is invalid", "Password is too short", "Password contains invalid characters"}
	if !reflect.DeepEqual(body.Errors, wantDetails) {
		t.Fatalf("legacy details: %v; want %v", body.Errors, wantDetails)
	}
}

func TestAPITransportErrorsUseTheSameEnvelope(t *testing.T) {
	h := router(t, &authStub{}, false, 30)
	for _, tc := range []struct {
		method, path, body, content string
		status                      int
		code                        string
	}{
		{"GET", "/api/missing", "", "", 404, "route_not_found"},
		{"GET", "/api/auth/login", "", "", 405, "method_not_allowed"},
		{"POST", "/api/auth/login", `{}`, "text/plain", 415, "unsupported_media_type"},
		{"POST", "/api/auth/login", `null`, "application/json", 400, "invalid_json"},
		{"POST", "/api/auth/login", `{"extra":true}`, "application/json", 400, "invalid_json"},
		{"POST", "/api/auth/login", `{} {}`, "application/json", 400, "invalid_json"},
		{"POST", "/api/auth/login", `{"email":"` + strings.Repeat("a", 17000) + `"}`, "application/json", 413, "request_too_large"},
		{"GET", "/api/auth/me", "", "", 401, "unauthenticated"},
	} {
		t.Run(tc.code+tc.body[:min(3, len(tc.body))], func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.content)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			requireAPIError(t, w, tc.status, tc.code)
			if tc.status == 405 && w.Header().Get("Allow") != "POST" {
				t.Fatal("Allow header lost")
			}
		})
	}
	w := request(characterRouter(t, &characterStub{}), "PUT", "/api/characters/active", `{}`, characterCookie())
	body := requireAPIError(t, w, 400, "invalid_request")
	if len(body.Fields["character_id"]) == 0 {
		t.Fatal("missing selection field error")
	}
}

func TestForbiddenOriginAndUnverifiedEmailHaveDifferentCodes(t *testing.T) {
	h := router(t, &authStub{err: auth.ErrUnverified}, false, 30)
	w := request(h, "POST", "/api/auth/login", `{"email":"a@example.test","password":"password"}`, nil)
	requireAPIError(t, w, 403, "email_unverified")
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Header.Set("Origin", "https://attacker.example")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	requireAPIError(t, w, 403, "origin_forbidden")
}

func TestRateLimitAndAuthenticationFailuresUseStableCodes(t *testing.T) {
	h := router(t, &authStub{}, false, 1)
	request(h, "POST", "/api/auth/login", `{}`, nil)
	w := request(h, "POST", "/api/auth/login", `{}`, nil)
	requireAPIError(t, w, 429, "rate_limited")
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After lost")
	}
	h = router(t, &authStub{err: errors.New("private-sql-detail")}, false, 30)
	w = request(h, "GET", "/api/auth/me", "", &http.Cookie{Name: "session_token", Value: "opaque-token"})
	requireAPIError(t, w, 500, "internal_error")
}
