package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest/internal/middleware"
	"gotest/internal/observability"
)

func TestObservationPreservesImplicitHeadersAndContext(t *testing.T) {
	var logs bytes.Buffer
	h := middleware.Observe(slog.New(slog.NewJSONHandler(&logs, nil)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := observability.RequestID(r.Context()); id == "" || id != w.Header().Get("X-Request-ID") {
			t.Fatal("request context lost the response ID")
		}
		observability.Logger(r.Context()).Info("handler_event")
		_, _ = w.Write([]byte("hello"))
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/example", nil))
	if w.Code != 200 || w.Body.String() != "hello" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("wrapper changed implicit response: %d %v %q", w.Code, w.Header(), w.Body.String())
	}
	decoder := json.NewDecoder(&logs)
	for {
		var record map[string]any
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if record["request_id"] != w.Header().Get("X-Request-ID") {
			t.Fatal("handler and access log do not share an ID")
		}
	}
}

func TestPanicAfterFlushAbortsWithoutAppendingErrorJSON(t *testing.T) {
	var logs bytes.Buffer
	h := middleware.Observe(slog.New(slog.NewJSONHandler(&logs, nil)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("partial"))
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Fatal(err)
		}
		panic("private-panic-content")
	}))
	w := httptest.NewRecorder()
	func() {
		defer func() {
			if recover() != http.ErrAbortHandler {
				t.Error("committed response must abort")
			}
		}()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/stream", nil))
	}()
	if !w.Flushed || w.Code != 200 || w.Body.String() != "partial" {
		t.Fatalf("panic corrupted committed response: %d %s", w.Code, w.Body)
	}
	if strings.Contains(logs.String(), "private-panic-content") {
		t.Fatal("panic value logged")
	}
	decoder := json.NewDecoder(&logs)
	var access map[string]any
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record["msg"] == "http_request" {
			access = record
		}
	}
	if access["status"] != float64(200) || access["bytes"] != float64(7) || access["aborted"] != true {
		t.Fatalf("aborted response was misreported: %+v", access)
	}
}

func TestAbortHandlerKeepsItsSemantics(t *testing.T) {
	var logs bytes.Buffer
	h := middleware.Observe(slog.New(slog.NewJSONHandler(&logs, nil)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(http.ErrAbortHandler) }))
	w := httptest.NewRecorder()
	func() {
		defer func() {
			if recover() != http.ErrAbortHandler {
				t.Error("intentional abort was swallowed")
			}
		}()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/abort", nil).WithContext(context.Background()))
	}()
	if w.Body.Len() != 0 || strings.Contains(logs.String(), "request_panic") {
		t.Fatal("intentional abort was turned into a JSON panic response")
	}
}
