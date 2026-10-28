package middleware

import (
	"bufio"
	"crypto/rand"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"gotest/internal/httpapi"
	"gotest/internal/observability"
)

// Observe wraps the whole router, including origin rejection and static files.
func Observe(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)
		r = r.WithContext(observability.WithRequest(r.Context(), logger, id))
		requestLog := observability.Logger(r.Context())
		response := &observedResponse{ResponseWriter: w}
		defer func() {
			aborted := false
			if value := recover(); value != nil {
				if value == http.ErrAbortHandler {
					aborted = true
				} else {
					requestLog.ErrorContext(r.Context(), "request_panic", "stack", string(debug.Stack()))
					if response.status == 0 {
						httpapi.WriteError(response, 500, httpapi.CodeInternalError, "Внутренняя ошибка сервера", nil)
					} else {
						aborted = true
					}
				}
			}
			status := response.status
			if status == 0 && !aborted {
				status = http.StatusOK
			}
			level := slog.LevelInfo
			if status >= 500 || aborted {
				level = slog.LevelError
			}
			written := response.bytes
			if r.Method == http.MethodHead {
				written = 0
			}
			requestLog.Log(r.Context(), level, "http_request", "method", r.Method,
				"path", r.URL.Path, "status", status, "bytes", written,
				"duration_ms", float64(time.Since(started))/float64(time.Millisecond), "aborted", aborted)
			if aborted {
				panic(http.ErrAbortHandler)
			}
		}()
		next.ServeHTTP(response, r)
	})
}

type observedResponse struct {
	http.ResponseWriter
	status, bytes int
}

func (w *observedResponse) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.ResponseWriter.WriteHeader(status)
	if status >= 200 || status == http.StatusSwitchingProtocols {
		w.status = status
	}
}

func (w *observedResponse) Write(body []byte) (int, error) {
	if w.status == 0 {
		// Let the underlying writer perform its implicit header/content-type logic.
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func (w *observedResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *observedResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *observedResponse) FlushError() error {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *observedResponse) Flush() { _ = w.FlushError() }
