package main

import (
	"gotest/internal/config"
	"gotest/internal/db"
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
	defer pool.Close()

	routes.Register()

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Listening on port %s", cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatal(err)
	}
}
