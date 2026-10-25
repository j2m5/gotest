package routes_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/observability"
	"gotest/internal/routes"
)

func probeRouter(t *testing.T, check func(context.Context) error, lifecycle ...context.Context) http.Handler {
	t.Helper()
	s := &authStub{}
	options := routes.Options{
		StaticDir: t.TempDir(), Ready: check, Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	if len(lifecycle) != 0 {
		options.Lifecycle = lifecycle[0]
	}
	h, err := routes.New(handlers.NewHandler(s, false), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(s), options)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestReadinessReflectsDependencyFailureAndRecovery(t *testing.T) {
	var dependencyError error
	calls := 0
	h := probeRouter(t, func(ctx context.Context) error {
		calls++
		if observability.RequestID(ctx) == "" {
			t.Error("dependency check lost request context")
		}
		return dependencyError
	})
	for _, available := range []bool{true, false, true} {
		dependencyError = nil
		if !available {
			dependencyError = errors.New("private-database-password")
		}
		w := request(h, "GET", "/readyz", "", nil)
		if available {
			if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
				t.Fatalf("ready response: %d %s", w.Code, w.Body)
			}
		} else {
			requireAPIError(t, w, 503, "not_ready")
			if strings.Contains(w.Body.String(), "private-") {
				t.Fatal("readiness disclosed dependency details")
			}
		}
	}
	if calls != 3 {
		t.Fatal("readiness cached an obsolete dependency status")
	}
	request(h, "GET", "/healthz", "", nil)
	if calls != 3 {
		t.Fatal("liveness performed a database check")
	}
	w := request(h, "POST", "/readyz", "", nil)
	requireAPIError(t, w, 405, "method_not_allowed")
	if calls != 3 {
		t.Fatal("unsupported method reached the database")
	}
}

func TestReadinessBoundsDatabaseWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := probeRouter(t, func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
		started := time.Now()
		w := request(h, "GET", "/readyz", "", nil)
		requireAPIError(t, w, 503, "not_ready")
		if elapsed := time.Since(started); elapsed <= 0 || elapsed > 3*time.Second {
			t.Fatalf("readiness did not enforce a short deadline: %v", elapsed)
		}
	})
}

func TestReadinessHonorsRequestCancellation(t *testing.T) {
	h := probeRouter(t, func(ctx context.Context) error { return ctx.Err() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil).WithContext(ctx))
	requireAPIError(t, w, 503, "not_ready")
}

func TestProbeHEADReturnsStatusWithoutBody(t *testing.T) {
	server := httptest.NewServer(probeRouter(t, func(context.Context) error { return nil }))
	defer server.Close()
	for _, path := range []string{"/healthz", "/readyz"} {
		response, err := server.Client().Head(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != 200 || len(body) != 0 || response.Header.Get("X-Request-ID") == "" || response.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("invalid HEAD probe: %s", path)
		}
	}
}

func TestReadinessCannotSucceedDuringShutdown(t *testing.T) {
	for _, when := range []string{"before", "during", "after"} {
		t.Run(when, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				lifecycle, stop := context.WithCancel(context.Background())
				defer stop()
				if when == "before" {
					stop()
				}
				h := probeRouter(t, func(probe context.Context) error {
					if when == "before" {
						t.Error("database probed after shutdown began")
					}
					stop()
					if when == "during" {
						<-probe.Done()
					}
					return nil // A successful/late dependency result must not override shutdown.
				}, lifecycle)
				started := time.Now()
				w := request(h, "GET", "/readyz", "", nil)
				requireAPIError(t, w, 503, "not_ready")
				if time.Since(started) >= time.Second {
					t.Fatal("shutdown waited for the ordinary probe timeout")
				}
			})
		})
	}
}
