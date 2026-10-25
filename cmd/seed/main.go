package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gotest/internal/db"
	"gotest/internal/seed"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	directory := flag.String("dir", "seeds", "directory containing SQL seed files")
	timeout := flag.Duration("timeout", time.Minute, "time limit for the entire seed operation")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected arguments; use -dir and -timeout")
	}
	if *timeout <= 0 {
		return errors.New("-timeout must be positive")
	}
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	dsn := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		return errors.New("DATABASE_URL is required")
	}

	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(interrupt, *timeout)
	defer cancel()
	connect, cancelConnect := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Open(connect, dsn)
	cancelConnect()
	if err != nil {
		return err
	}
	defer pool.Close()
	files, err := seed.Run(ctx, pool, *directory)
	if err != nil {
		return err
	}
	for _, file := range files {
		log.Printf("Seeded: %s", file)
	}
	log.Printf("Done: %d files committed", len(files))
	return nil
}
