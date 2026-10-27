package db_test

import (
	"context"
	"testing"
	"time"

	"gotest/internal/auth"
)

// Fixtures supply the current stored credential just as a successful login does.
func createStoredSession(t *testing.T, s *testStorage, ctx context.Context, userID int, token string, expires time.Time) error {
	t.Helper()
	var password string
	if err := s.pool.QueryRow(ctx, `SELECT password FROM users WHERE id=$1`, userID).Scan(&password); err != nil {
		t.Fatal(err)
	}
	return s.CreateSession(ctx, auth.SessionInput{UserID: userID, PasswordHash: password, TokenHash: token, ExpiresAt: expires})
}
