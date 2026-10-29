package routes

import (
	"gotest/internal/handlers"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
	"net/http"
)

func registerForumRoutes(api *http.ServeMux, h *handlers.ForumHandler, m *middleware.Middleware) {
	var categories, topics, topic, create, reply, editTopic, editPost, moderate http.HandlerFunc
	if h != nil {
		categories, topics, topic = h.Categories, h.Topics, h.Topic
		create, reply, editTopic, editPost, moderate = h.Create, h.Reply, h.EditTopic, h.EditPost, h.Moderate
	}
	unavailable := func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, 503, "not_ready", "Форум недоступен", nil)
	}
	for path, methods := range map[string]map[string]http.HandlerFunc{
		"/api/forum/categories": {"GET": categories},
		"/api/forum/topics":     {"GET": topics, "POST": create},
		"/api/forum/topic":      {"GET": topic},
		"/api/forum/replies":    {"POST": reply},
		"/api/forum/edit-topic": {"POST": editTopic},
		"/api/forum/edit-reply": {"POST": editPost},
		"/api/forum/moderate":   {"POST": moderate},
	} {
		for method, handler := range methods {
			if handler == nil {
				handler = unavailable
			}
			methods[method] = m.Auth(handler)
		}
		api.Handle(path, methodHandler(methods))
	}
}
