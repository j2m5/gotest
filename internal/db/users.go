package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gotest/internal/auth"
	"gotest/internal/models"
)

var _ auth.Repository = (*Storage)(nil)

func authStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrNotFound
	}
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) && postgres.Code == "23505" {
		return auth.ErrConflict
	}
	return err
}

func (s *Storage) CreateRegistration(ctx context.Context, email, login, passwordHash, tokenHash string, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID int
	err = tx.QueryRow(ctx, `INSERT INTO users(email,login,password) VALUES($1,$2,$3) RETURNING id`, email, login, passwordHash).Scan(&userID)
	if err != nil {
		return authStorageError(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO email_verifications(user_id,token,expired_at) VALUES($1,$2,$3)`, userID, tokenHash, expiresAt); err != nil {
		return authStorageError(err)
	}
	return authStorageError(tx.Commit(ctx))
}

func (s *Storage) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `SELECT id,role_id,character_id,name,email,login,password,email_verified_at,created_at,updated_at FROM users WHERE lower(btrim(email))=lower(btrim($1))`, email).Scan(
		&u.ID, &u.RoleID, &u.CharacterID, &u.Name, &u.Email, &u.Login, &u.Password, &u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, authStorageError(err)
	}
	return u, nil
}
