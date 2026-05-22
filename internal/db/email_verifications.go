package db

import (
	"context"
	"gotest/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateEmailVerification(pool *pgxpool.Pool, userID int, token string) error {
	query := "INSERT INTO email_verifications (user_id, token) VALUES ($1, $2)"

	_, err := pool.Exec(context.Background(), query, userID, token)

	return err
}

func DeleteEmailVerification(pool *pgxpool.Pool, token string) error {
	query := "DELETE FROM email_verifications WHERE token = $1"

	_, err := pool.Exec(context.Background(), query, token)

	return err
}

func FindEmailVerificationByToken(pool *pgxpool.Pool, token string) (*models.EmailVerification, error) {
	query := "SELECT id, user_id, token, expired_at, created_at, updated_at FROM email_verifications WHERE token = $1"

	row := pool.QueryRow(context.Background(), query, token)

	emailVerification := &models.EmailVerification{}

	err := row.Scan(&emailVerification.ID, &emailVerification.UserID, &emailVerification.Token, &emailVerification.ExpiredAt, &emailVerification.CreatedAt, &emailVerification.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return emailVerification, nil
}
