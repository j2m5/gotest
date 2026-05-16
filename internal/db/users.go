package db

import (
	"context"
	"gotest/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, email string, login string, password string) error {
	query := "INSERT INTO users (email, login, password) VALUES ($1, $2, $3)"

	_, err := pool.Exec(context.Background(), query, email, login, password)

	return err
}

func FindUserByEmail(pool *pgxpool.Pool, email string) (*models.User, error) {
	query := `
		SELECT id, role_id, name, email, login, password, email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := pool.QueryRow(context.Background(), query, email)

	user := &models.User{}

	err := row.Scan(&user.ID, &user.RoleID, &user.Name, &user.Email, &user.Login, &user.Password, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
