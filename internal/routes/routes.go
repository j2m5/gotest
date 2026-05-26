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
}
