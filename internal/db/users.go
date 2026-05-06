package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, username string, password string, role string) error {
	query := "INSERT INTO users (username, password, role) VALUES ($1, $2, $3)"

	_, err := pool.Exec(context.Background(), query, username, password, role)

	return err
}
