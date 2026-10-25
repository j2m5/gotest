package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gotest/internal/httpapi"
	"gotest/internal/observability"
)

func Health(w http.ResponseWriter, r *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func Readiness(check func(context.Context) error, lifecycle context.Context) http.HandlerFunc {
	if lifecycle == nil {
		lifecycle = context.Background()
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			stop := context.AfterFunc(lifecycle, cancel)
			defer stop()
			err := lifecycle.Err()
			if err == nil {
				err = ctx.Err()
			}
			if err == nil {
				err = check(ctx)
			}
			if err == nil {
				err = ctx.Err()
			}
			// AfterFunc runs asynchronously; recheck the lifecycle synchronously too.
			if err == nil {
				err = lifecycle.Err()
			}
			if err == nil {
				Health(w, r)
				return
			}
			observability.Logger(r.Context()).LogAttrs(r.Context(), slog.LevelWarn, "readiness_failed", observability.ErrorAttrs(err)...)
		}
		httpapi.WriteError(w, http.StatusServiceUnavailable, "not_ready", "Сервис временно недоступен", nil)
	}
}
