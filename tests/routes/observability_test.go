package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"gotest/internal/auth"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &output
}

func logRecords(t *testing.T, output *bytes.Buffer) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	var records []map[string]any
	for {
		var record map[string]any
		if err := decoder.Decode(&record); err == io.EOF {
			return records
		} else if err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		records = append(records, record)
	}
}

func TestRequestsHaveUniqueServerIDs(t *testing.T) {
	captureLogs(t)
	h := router(t, &authStub{}, false, 30)
	seen := map[string]bool{}
	for _, path := range []string{"/api/auth/me", "/api/missing", "/assets/app.js", "/verify?token=private-token"} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-Request-ID", "client-chosen-id")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		id := w.Header().Get("X-Request-ID")
		if len(id) < 20 || id == "client-chosen-id" || seen[id] || strings.ContainsAny(id, "\r\n\t ") {
			t.Fatalf("missing, reused or untrusted request ID for %s: %q", path, id)
		}
		seen[id] = true
	}
}

func TestRequestFailureLogsAreCorrelatedAndExcludeSecrets(t *testing.T) {
	output := captureLogs(t)
	h := router(t, &authStub{err: errors.New("private-error-password")}, false, 30)
	r := httptest.NewRequest("POST", "/api/auth/login?token=private-query-token", strings.NewReader(`{"email":"private-email@example.test","password":"private-body-password"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer private-header-token")
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "private-cookie-token"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	requireAPIError(t, w, 500, "internal_error")
	if strings.Contains(output.String(), "private-") {
		t.Fatal("request secrets or raw dependency error escaped into logs")
	}
	records := logRecords(t, output)
	var access, failure map[string]any
	for _, record := range records {
		if record["request_id"] != w.Header().Get("X-Request-ID") || record["request_id"] == "" {
			t.Fatalf("uncorrelated record: %+v", record)
		}
		if record["msg"] == "http_request" {
			access = record
		} else if record["level"] == "ERROR" {
			failure = record
		}
	}
	if access == nil || failure == nil || access["method"] != "POST" || access["path"] != "/api/auth/login" || access["status"] != float64(500) || access["bytes"] != float64(w.Body.Len()) {
		t.Fatalf("missing request/error diagnostics: %+v", records)
	}
	if duration, ok := access["duration_ms"].(float64); !ok || duration < 0 {
		t.Fatalf("invalid request duration: %+v", access)
	}
}

func TestHealthIsPublicWithoutDatabaseOrSPABuild(t *testing.T) {
	captureLogs(t)
	s := &authStub{err: errors.New("database unavailable")}
	h, err := routes.New(handlers.NewHandler(s, false), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(s), routes.Options{StaticDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	w := request(h, "GET", "/healthz", "", nil)
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` || w.Header().Get("Cache-Control") != "no-store" || s.calls != 0 {
		t.Fatalf("liveness depends on application dependencies: %d %s", w.Code, w.Body)
	}
	w = request(h, "POST", "/healthz", "", nil)
	requireAPIError(t, w, 405, "method_not_allowed")
}

func TestReadinessWithoutDependencyCheckIsUnavailable(t *testing.T) {
	captureLogs(t)
	w := request(router(t, &authStub{}, false, 30), "GET", "/readyz", "", nil)
	requireAPIError(t, w, 503, "not_ready")
}

type panickingAuth struct{ *authStub }

func (s *panickingAuth) Login(context.Context, string, string) (*auth.LoginResult, error) {
	panic("private-panic-value")
}

func TestPanicHasSafeJSONAndCorrelatedDiagnostics(t *testing.T) {
	output := captureLogs(t)
	s := &panickingAuth{&authStub{}}
	h, err := routes.New(handlers.NewHandler(s, false), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(s), routes.Options{StaticDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	w := request(h, "POST", "/api/auth/login", `{}`, nil)
	requireAPIError(t, w, 500, "internal_error")
	if strings.Contains(output.String()+w.Body.String(), "private-panic-value") {
		t.Fatal("panic value disclosed")
	}
	var panicLog bool
	for _, record := range logRecords(t, output) {
		if record["msg"] == "request_panic" && record["request_id"] == w.Header().Get("X-Request-ID") && record["stack"] != "" {
			panicLog = true
		}
	}
	if !panicLog {
		t.Fatal("panic has no correlated diagnostic record")
	}
}

func TestDependencyDiagnosticsKeepCategoriesWithoutRawErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		kind, state string
	}{
		{"postgres", fmt.Errorf("private-wrapper: %w", &pgconn.PgError{Code: "23505", Message: "private-sql-message", Detail: "private-email@example.test"}), "database", "23505"},
		{"deadline", fmt.Errorf("private-timeout: %w", context.DeadlineExceeded), "timeout", ""},
		{"cancel", fmt.Errorf("private-cancel: %w", context.Canceled), "canceled", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := captureLogs(t)
			w := request(router(t, &authStub{err: tc.err}, false, 30), "POST", "/api/auth/login", `{}`, nil)
			requireAPIError(t, w, 500, "internal_error")
			if strings.Contains(output.String(), "private-") {
				t.Fatal("raw dependency diagnostics leaked")
			}
			found := false
			for _, record := range logRecords(t, output) {
				if record["msg"] == "request_failed" {
					found = record["error_kind"] == tc.kind && (tc.state == "" || record["sqlstate"] == tc.state)
				}
			}
			if !found {
				t.Fatal("wrapped error lost its diagnostic category")
			}
		})
	}
}

func TestRejectedRequestsKeepAccessLogsAndIDs(t *testing.T) {
	output := captureLogs(t)
	h := router(t, &authStub{}, false, 1)
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Header.Set("Origin", "https://untrusted.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	requireAPIError(t, w, 403, "origin_forbidden")
	request(h, "POST", "/api/auth/login", `{}`, nil)
	w = request(h, "POST", "/api/auth/login", `{}`, nil)
	requireAPIError(t, w, 429, "rate_limited")
	var statuses []int
	for _, record := range logRecords(t, output) {
		if record["msg"] == "http_request" {
			if record["request_id"] == nil || record["request_id"] == "" {
				t.Fatal("rejected request has no ID")
			}
			statuses = append(statuses, int(record["status"].(float64)))
		}
	}
	if len(statuses) != 3 || statuses[0] != 403 || statuses[1] != 200 || statuses[2] != 429 {
		t.Fatalf("missing rejected requests in access log: %v", statuses)
	}
}
