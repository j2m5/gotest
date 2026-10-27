package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/auth"
	"gotest/internal/notification"
)

// EnqueuePasswordReset deliberately does not look up the account. The public
// acceptance path is identical for known and unknown email addresses.
func (s *AuthRepository) EnqueuePasswordReset(ctx context.Context, email string, expiresAt time.Time) error {
	message, err := s.cipher.Seal(notification.ResetRequest, email, "", expiresAt)
	if err != nil {
		return err
	}
	return NewMailQueue(s.pool).EnqueueResetRequest(ctx, message)
}

// CreatePasswordReset serializes replacement and a per-account cooldown across
// application instances. Unknown accounts and throttled requests have no effects.
func (s *AuthRepository) CreatePasswordReset(ctx context.Context, input auth.PasswordReset) (bool, error) {
	email, tokenHash, expiresAt := input.Email, input.TokenHash, input.ExpiresAt
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	// Deleting the owned input inside this transaction makes reprocessing
	// idempotent. A failure below rolls this deletion back with all other writes.
	if err := consumeResetRequest(ctx, tx, input.RequestID, input.RequestLease); err != nil {
		return false, err
	}
	var userID int
	err = tx.QueryRow(ctx, `SELECT id,email FROM users WHERE lower(btrim(email))=lower(btrim($1)) FOR UPDATE`, email).Scan(&userID, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}
	if err != nil {
		return false, err
	}
	var coolingDown bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM password_resets WHERE user_id=$1 AND created_at>clock_timestamp()-interval '1 minute')`, userID).Scan(&coolingDown); err != nil {
		return false, err
	}
	if coolingDown {
		return false, tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `INSERT INTO password_resets(user_id,token,expires_at) VALUES($1,$2,$3)
		ON CONFLICT (user_id) DO UPDATE SET token=EXCLUDED.token,expires_at=EXCLUDED.expires_at,created_at=clock_timestamp()`, userID, tokenHash, expiresAt)
	if err != nil {
		return false, authStorageError(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM mail_jobs WHERE user_id=$1 AND kind='password_reset'`, userID); err != nil {
		return false, err
	}
	message, err := s.cipher.Seal(notification.PasswordReset, email, input.Token, expiresAt)
	if err != nil {
		return false, err
	}
	if err := enqueueMail(ctx, tx, message, &userID); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, authStorageError(err)
	}
	return true, nil
}

// ResetPassword uses the same user-first lock order as login/logout/character
// mutations. It rechecks the token and the database clock after any lock wait.
func (s *AuthRepository) ResetPassword(ctx context.Context, tokenHash, passwordHash string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var userID int
	err = tx.QueryRow(ctx, `SELECT user_id FROM password_resets WHERE token=$1`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrInvalidResetToken
	}
	if err != nil {
		return "", err
	}
	var email string
	err = tx.QueryRow(ctx, `SELECT email FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrInvalidResetToken
	}
	if err != nil {
		return "", err
	}
	var valid bool
	err = tx.QueryRow(ctx, `SELECT expires_at>clock_timestamp() FROM password_resets WHERE token=$1 AND user_id=$2 FOR UPDATE`, tokenHash, userID).Scan(&valid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrInvalidResetToken
	}
	if err != nil {
		return "", err
	}
	if !valid {
		return "", auth.ErrInvalidResetToken
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET password=$2,character_id=NULL WHERE id=$1`, userID, passwordHash); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, userID); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM password_resets WHERE user_id=$1`, userID); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM mail_jobs WHERE user_id=$1 AND kind='password_reset'`, userID); err != nil {
		return "", err
	}
	message, err := s.cipher.Seal(notification.PasswordChanged, email, "", time.Now().Add(24*time.Hour))
	if err != nil {
		return "", err
	}
	if err := enqueueMail(ctx, tx, message, &userID); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", authStorageError(err)
	}
	return email, nil
}
