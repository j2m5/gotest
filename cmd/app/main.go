package main

import (
	"gotest/internal/db"
	"log"
	"net/http"
)

func main() {
	pool := db.New()

	defer pool.Close()

	err := db.CreateUser(pool, "admin", "admin", "admin")

	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GOTest server running"))
	})

	log.Println("Listening on port 8081")

	err = http.ListenAndServe(":8081", nil)

	if err != nil {
		log.Fatal(err)
	}
}
