package routes

import (
	"net/http"

	"gotest/internal/handlers"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
)

func registerCombatRoutes(api *http.ServeMux, h *handlers.CombatHandler, m *middleware.Middleware) {
	unavailable := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, 503, "not_ready", "Данные боя недоступны", nil)
	})
	current, start, use, retreat, detail := unavailable, unavailable, unavailable, unavailable, unavailable
	if h != nil {
		current, start, use, retreat, detail = h.Current, h.Start, h.Use, h.Retreat, h.View
	}
	for _, endpoint := range []struct {
		path, method string
		handler      http.HandlerFunc
	}{
		{"/api/battles/current", "GET", current}, {"/api/battles/detail", "GET", detail},
		{"/api/battles/start", "POST", start}, {"/api/battles/use", "POST", use}, {"/api/battles/retreat", "POST", retreat},
	} {
		api.Handle(endpoint.path, methodHandler(map[string]http.HandlerFunc{endpoint.method: m.Auth(endpoint.handler)}))
	}
}
