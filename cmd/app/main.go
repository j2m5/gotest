package main

import (
	"gotest/internal/auth"
	"gotest/internal/config"
	"gotest/internal/db"
	"gotest/internal/flash"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	pool := db.New(cfg.DatabaseURL)
	storage := db.NewStorage(pool)
	a := auth.NewAuth(storage)
	m := middleware.NewMiddleware(a)
	f := flash.NewFlash(cfg.FlashSecret)
	handler := handlers.NewHandler(storage, a, f)
	defer pool.Close()

	routes.Register(handler, m)

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Listening on port %s", cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatal(err)
	}
}
