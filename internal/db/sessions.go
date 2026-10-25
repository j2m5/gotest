package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/auth"
)

func (s *AuthRepository) CreateSession(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO sessions(user_id,token,expires_at) VALUES($1,$2,$3)`, userID, tokenHash, expiresAt)
	return authStorageError(err)
}

func (s *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	var userID int
	err = tx.QueryRow(ctx, `SELECT user_id FROM sessions WHERE token=$1`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	// Serialize account changes using the same user-first order as verification.
	// A token read before this lock can have been revoked while we waited.
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // Account deletion has already cascaded to its sessions.
	}
	if err != nil {
		return err
	}
	var active bool
	err = tx.QueryRow(ctx, `DELETE FROM sessions WHERE token=$1 AND user_id=$2 RETURNING expires_at>clock_timestamp()`, tokenHash, userID).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // A repeated logout must not clear a later selection.
	}
	if err != nil {
		return err
	}
	if active {
		if _, err := tx.Exec(ctx, `UPDATE users SET character_id=NULL WHERE id=$1 AND character_id IS NOT NULL`, userID); err != nil {
			return err
		}
	}
	return authStorageError(tx.Commit(ctx))
}

func (s *AuthRepository) FindSession(ctx context.Context, tokenHash string) (*auth.Session, error) {
	u := &auth.Identity{}
	session := &auth.Session{User: u}
	err := s.pool.QueryRow(ctx, `SELECT u.id,u.role_id,u.character_id,u.email,u.login,u.email_verified_at,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token=$1`, tokenHash).Scan(
		&u.ID, &u.RoleID, &u.CharacterID, &u.Email, &u.Login, &u.EmailVerifiedAt, &session.ExpiresAt,
	)
	if err != nil {
		return nil, authStorageError(err)
	}
	return session, nil
}
