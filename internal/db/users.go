package db

import (
	"context"
	"gotest/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, username string, password string, role string) error {
	query := "INSERT INTO users (username, password, role) VALUES ($1, $2, $3)"

	_, err := pool.Exec(context.Background(), query, username, password, role)

	return err
}

func FindUserByUsername(pool *pgxpool.Pool, username string) (*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	row := pool.QueryRow(context.Background(), query, username)

	user := &models.User{}

	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
