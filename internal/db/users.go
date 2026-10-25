package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gotest/internal/auth"
)

var _ auth.Repository = (*AuthRepository)(nil)

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

func (s *AuthRepository) CreateRegistration(ctx context.Context, registration auth.Registration) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID int
	err = tx.QueryRow(ctx, `INSERT INTO users(email,login,password) VALUES($1,$2,$3) RETURNING id`, registration.Email, registration.Login, registration.PasswordHash).Scan(&userID)
	if err != nil {
		return authStorageError(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO email_verifications(user_id,token,expired_at) VALUES($1,$2,$3)`, userID, registration.VerificationTokenHash, registration.VerificationExpiresAt); err != nil {
		return authStorageError(err)
	}
	return authStorageError(tx.Commit(ctx))
}

func (s *AuthRepository) FindCredentialsByEmail(ctx context.Context, email string) (*auth.Credentials, error) {
	credentials := &auth.Credentials{}
	u := &credentials.User
	err := s.pool.QueryRow(ctx, `SELECT id,role_id,character_id,email,login,password,email_verified_at FROM users WHERE lower(btrim(email))=lower(btrim($1))`, email).Scan(
		&u.ID, &u.RoleID, &u.CharacterID, &u.Email, &u.Login, &credentials.PasswordHash, &u.EmailVerifiedAt,
	)
	if err != nil {
		return nil, authStorageError(err)
	}
	return credentials, nil
}
