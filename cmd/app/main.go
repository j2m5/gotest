package main

import (
	"gotest/internal/db"
	"gotest/internal/routes"
	"log"
	"net/http"
)

func main() {
	pool := db.New()

	defer pool.Close()

	routes.Register()

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Listening on port 8081")

	err := http.ListenAndServe(":8081", nil)

	if err != nil {
		log.Fatal(err)
	}
}
