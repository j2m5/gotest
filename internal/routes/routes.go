package routes

import (
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"net/http"
)

func Register(h *handlers.Handler, m *middleware.Middleware) {
	http.HandleFunc("/", m.Auth(h.Home))
	http.HandleFunc("/register", m.Guest(h.Register))
	http.HandleFunc("/login", m.Guest(h.Login))
	http.HandleFunc("/logout", m.Auth(h.Logout))
	http.HandleFunc("/verify", h.VerifyEmail)

	http.HandleFunc("/api/auth/register", h.Register)
	http.HandleFunc("/api/auth/login", h.Login)
	http.HandleFunc("/api/auth/logout", m.Auth(h.Logout))
	http.HandleFunc("/api/auth/verify", m.Auth(h.VerifyEmail))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})
}
