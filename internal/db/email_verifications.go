package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/auth"
)

func (s *Storage) ConsumeVerification(ctx context.Context, tokenHash string, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Read only the owner here. Both verification and resend lock users first;
	// after acquiring that lock read the token again to detect replacement.
	var userID int
	err = tx.QueryRow(ctx, `SELECT user_id FROM email_verifications WHERE token=$1`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrInvalidToken
	}
	if err != nil {
		return err
	}
	var verifiedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&verifiedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrInvalidToken
	}
	if err != nil {
		return err
	}
	if verifiedAt != nil {
		return auth.ErrInvalidToken
	}
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `SELECT expired_at FROM email_verifications WHERE token=$1 AND user_id=$2 FOR UPDATE`, tokenHash, userID).Scan(&expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrInvalidToken
	}
	if err != nil {
		return err
	}
	if !expiresAt.After(now) {
		return auth.ErrInvalidToken
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET email_verified_at=$2,updated_at=$2 WHERE id=$1`, userID, now.UTC()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM email_verifications WHERE user_id=$1`, userID); err != nil {
		return err
	}
	return authStorageError(tx.Commit(ctx))
}

func (s *Storage) ReplaceVerification(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var verifiedAt *time.Time
	if err = tx.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&verifiedAt); err != nil {
		return authStorageError(err)
	}
	if verifiedAt != nil {
		return auth.ErrAlreadyVerified
	}
	if _, err = tx.Exec(ctx, `DELETE FROM email_verifications WHERE user_id=$1`, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO email_verifications(user_id,token,expired_at) VALUES($1,$2,$3)`, userID, tokenHash, expiresAt); err != nil {
		return authStorageError(err)
	}
	return authStorageError(tx.Commit(ctx))
}
