package db_test

import (
	"testing"

	"gotest/internal/auth"
	"gotest/internal/maintenance"
)

func TestCleanupIncludesPasswordResets(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for i, email := range []string{"expired-reset@example.test", "live-reset@example.test"} {
		var userID int
		if err := s.pool.QueryRow(ctx, `INSERT INTO users(email,login,password) VALUES($1,$1,'unused') RETURNING id`, email).Scan(&userID); err != nil {
			t.Fatal(err)
		}
		interval := "-1 hour"
		if i == 1 {
			interval = "1 hour"
		}
		mustExec(t, ctx, s, `INSERT INTO password_resets(user_id,token,expires_at) VALUES($1,$2,clock_timestamp()+$3::interval)`, userID, auth.TokenDigest(email), interval)
	}
	preview, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1, DryRun: true})
	if err != nil || preview.PasswordResets != 1 {
		t.Fatalf("preview: %+v %v", preview, err)
	}
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || result.PasswordResets != 1 {
		t.Fatalf("cleanup: %+v %v", result, err)
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets WHERE expires_at>clock_timestamp()`).Scan(&count); err != nil || count != 1 {
		t.Fatal("live reset removed")
	}
	again, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || again.PasswordResets != 0 {
		t.Fatalf("repeat cleanup: %+v %v", again, err)
	}
}
