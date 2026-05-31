package main

import (
	"context"
	"gotest/internal/config"
	"gotest/internal/db"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	pool := db.New(cfg.DatabaseURL)
	defer pool.Close()

	files, err := filepath.Glob("seeds/*.sql")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Cannot read file %s: %v", file, err)
		}

		_, err = pool.Exec(context.Background(), string(sql))
		if err != nil {
			log.Fatalf("Cannot seed %s: %v", file, err)
		}

		log.Printf("Seeded: %s", file)
	}

	log.Println("Done!")
}
