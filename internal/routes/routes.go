package routes

import (
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"net/http"
)

func Register() {
	http.HandleFunc("/", middleware.Auth(handlers.Home))
	http.HandleFunc("/register", middleware.Guest(handlers.Register))
	http.HandleFunc("/login", middleware.Guest(handlers.Login))
	http.HandleFunc("/logout", middleware.Auth(handlers.Logout))
	http.HandleFunc("/verify", middleware.Guest(handlers.VerifyEmail))
}
