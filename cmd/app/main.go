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

	log.Println("Listening on port 8081")

	err := http.ListenAndServe(":8081", nil)

	if err != nil {
		log.Fatal(err)
	}
}
