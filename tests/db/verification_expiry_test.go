package db_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/db"
)

// Expiry must be evaluated after waiting for the account lock.
func TestVerificationRechecksExpiryAfterLockWait(t *testing.T) {
	for _, lock := range []string{"account", "token"} {
		t.Run(lock, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			hash := auth.TokenDigest("audit-verification-expiry")
			mustExec(t, ctx, s, `UPDATE users SET email_verified_at=NULL WHERE id=1`)
			mustExec(t, ctx, s, `INSERT INTO email_verifications(user_id,token,expired_at) VALUES(1,$1,clock_timestamp()+interval '500 milliseconds')`, hash)
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
			var pid int
			query := "SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE"
			if lock == "token" {
				query = "SELECT pg_backend_pid() FROM email_verifications WHERE user_id=1 FOR UPDATE"
			}
			if err := tx.QueryRow(ctx, query).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- s.ConsumeVerification(ctx, hash) }()
			awaitResetUserLock(t, ctx, s, pid, done)
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				var expired bool
				if err := s.pool.QueryRow(ctx, `SELECT expired_at<=clock_timestamp() FROM email_verifications WHERE token=$1`, hash).Scan(&expired); err != nil {
					t.Fatal(err)
				}
				if expired {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-ticker.C:
				}
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			err = <-done
			var verified bool
			if queryErr := s.pool.QueryRow(ctx, `SELECT email_verified_at IS NOT NULL FROM users WHERE id=1`).Scan(&verified); queryErr != nil {
				t.Fatal(queryErr)
			}
			if !errors.Is(err, auth.ErrInvalidToken) || verified {
				t.Fatalf("expired verification was accepted after lock wait: err=%v verified=%v", err, verified)
			}
		})
	}
}

// A schema-local clock fixes the exact boundary without a timing tolerance.
// Other tests use the real PostgreSQL wall clock and real row-lock waits.
func TestVerificationExactDatabaseExpiryBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, expiry string
		valid        bool
	}{
		{"before", "2040-01-01 11:59:59+00", false},
		{"equal", "2040-01-01 12:00:00+00", false},
		{"after", "2040-01-01 12:00:00.000001+00", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base, ctx := characterTestStorage(t, true)
			mustExec(t, ctx, base, "CREATE FUNCTION clock_timestamp() RETURNS timestamptz LANGUAGE SQL VOLATILE AS $$ SELECT '2040-01-01 12:00:00+00'::timestamptz $$")
			config := base.pool.Config()
			config.ConnConfig.RuntimeParams["search_path"] += ",pg_catalog"
			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(pool.Close)
			s := &testStorage{AuthRepository: db.NewAuthRepository(pool, testCipher()), pool: pool}
			mustExec(t, ctx, s, "UPDATE users SET email_verified_at=NULL WHERE id=1")
			if err := s.ReplaceVerification(ctx, 1, "boundary", mustParseVerificationTime(t, "2041-01-01 12:00:00+00")); err != nil {
				t.Fatal(err)
			}
			mustExec(t, ctx, s, "UPDATE email_verifications SET expired_at=$1 WHERE token='boundary'", mustParseVerificationTime(t, tc.expiry))
			err = s.ConsumeVerification(ctx, "boundary")
			if tc.valid && err != nil || !tc.valid && !errors.Is(err, auth.ErrInvalidToken) {
				t.Fatalf("boundary valid=%v err=%v", tc.valid, err)
			}
			var intact bool
			if err := pool.QueryRow(ctx, "SELECT email_verified_at IS NULL AND EXISTS(SELECT 1 FROM email_verifications WHERE user_id=1) AND EXISTS(SELECT 1 FROM mail_jobs WHERE user_id=1 AND kind='verification') FROM users WHERE id=1").Scan(&intact); err != nil {
				t.Fatal(err)
			}
			if intact == tc.valid {
				t.Fatal("token, verification state and notification must change atomically")
			}
			if tc.valid {
				var complete bool
				if err := pool.QueryRow(ctx, "SELECT email_verified_at='2040-01-01 12:00:00+00'::timestamptz AND NOT EXISTS(SELECT 1 FROM email_verifications WHERE user_id=1) AND NOT EXISTS(SELECT 1 FROM mail_jobs WHERE user_id=1 AND kind='verification') FROM users WHERE id=1").Scan(&complete); err != nil || !complete {
					t.Fatalf("database time/consumption: %v %v", complete, err)
				}
				if err := s.ConsumeVerification(ctx, "boundary"); !errors.Is(err, auth.ErrInvalidToken) {
					t.Fatalf("reused token: %v", err)
				}
			}
		})
	}
}

func mustParseVerificationTime(t *testing.T, value string) time.Time {
	t.Helper()
	result, err := time.Parse("2006-01-02 15:04:05.999999Z07:00", strings.ReplaceAll(value, "+00", "+00:00"))
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestVerificationCancellationReleasesConnectionAndAllowsRetry(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, "UPDATE users SET email_verified_at=NULL WHERE id=1")
	if err := s.ReplaceVerification(ctx, 1, "cancelled", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	tx, pid := lockResetUser(t, ctx, s)
	cancelledCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.ConsumeVerification(cancelledCtx, "cancelled") }()
	awaitResetUserLock(t, ctx, s, pid, done)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation: %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	// pgx destroys cancelled connections asynchronously; wait for pool release.
	releaseCtx, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for s.pool.Stat().AcquiredConns() != 0 {
		select {
		case <-releaseCtx.Done():
			t.Fatal("connection retained after cancellation")
		case <-ticker.C:
		}
	}
	if err := s.ConsumeVerification(ctx, "cancelled"); err != nil {
		t.Fatalf("retry after cancellation: %v", err)
	}
}
