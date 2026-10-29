package routes

import (
	"net/http"

	"gotest/internal/handlers"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
)

func registerPresenceRoutes(api *http.ServeMux, h *handlers.PresenceHandler, m *middleware.Middleware) {
	unavailable := func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, 503, "not_ready", "Статус онлайна временно недоступен", nil)
	}
	snapshot, online, heartbeat := http.HandlerFunc(unavailable), http.HandlerFunc(unavailable), http.HandlerFunc(unavailable)
	if h != nil {
		snapshot, online, heartbeat = h.Snapshot, h.Online, h.Heartbeat
	}
	api.Handle("/api/presence", methodHandler(map[string]http.HandlerFunc{"GET": m.Auth(snapshot)}))
	api.Handle("/api/presence/online", methodHandler(map[string]http.HandlerFunc{"GET": m.Auth(online)}))
	api.Handle("/api/presence/heartbeat", methodHandler(map[string]http.HandlerFunc{"POST": m.Auth(heartbeat)}))
}
