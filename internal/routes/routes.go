package routes

import (
	"gotest/internal/handlers"
	"net/http"
)

func Register() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/logout", handlers.Logout)
}
