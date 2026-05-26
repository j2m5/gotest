package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func New(connStr string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		log.Fatal(err)
	}

	err = pool.Ping(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to PostgreSQL")

	Pool = pool

	return pool
}
