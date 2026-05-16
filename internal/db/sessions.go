package db

import (
	"context"
	"gotest/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateSession(pool *pgxpool.Pool, userID int, token string) error {
	query := "INSERT INTO sessions (user_id, token) VALUES ($1, $2)"

	_, err := pool.Exec(context.Background(), query, userID, token)

	return err
}

func DeleteSession(pool *pgxpool.Pool, token string) error {
	query := "DELETE FROM sessions WHERE token = $1"

	_, err := pool.Exec(context.Background(), query, token)

	return err
}

func FindUserBySessionToken(pool *pgxpool.Pool, token string) (*models.User, error) {
	query := `
		SELECT
		    users.id,
			users.role_id,
			users.name,
			users.email,
			users.login,
			users.password,
			users.email_verified_at,
			users.created_at,
			users.updated_at
		FROM sessions
		JOIN users
		ON users.id = sessions.user_id
		
		WHERE sessions.token = $1
	`

	row := pool.QueryRow(context.Background(), query, token)

	user := &models.User{}

	err := row.Scan(&user.ID, &user.RoleID, &user.Name, &user.Email, &user.Login, &user.Password, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
