package main

import (
	"gotest/internal/db"
	"gotest/internal/handlers"
	"log"
	"net/http"
)

func main() {
	pool := db.New()

	defer pool.Close()

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/register", handlers.Register)

	log.Println("Listening on port 8081")

	err := http.ListenAndServe(":8081", nil)

	if err != nil {
		log.Fatal(err)
	}
}
