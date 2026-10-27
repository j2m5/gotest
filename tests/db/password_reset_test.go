package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/auth"
)

func createResetLink(t *testing.T, ctx context.Context, s *testStorage, email, raw string) string {
	t.Helper()
	hash := auth.TokenDigest(raw)
	created, err := s.CreatePasswordReset(ctx, email, hash, time.Now().Add(time.Hour))
	if err != nil || !created {
		t.Fatalf("create reset: created=%v err=%v", created, err)
	}
	return hash
}

func TestPasswordResetLifecycleAndAccountIsolation(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, token := range []string{"ordinary", "remembered", "expired"} {
		if err := createStoredSession(t, s, ctx, 1, auth.TokenDigest(token), time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	if err := createStoredSession(t, s, ctx, 2, auth.TokenDigest("other-session"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=CASE id WHEN 1 THEN 1 ELSE 3 END`)
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, auth.TokenDigest("expired"))
	var verified time.Time
	if err := s.pool.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id=1`).Scan(&verified); err != nil {
		t.Fatal(err)
	}
	if created, err := s.CreatePasswordReset(ctx, "missing@example.test", auth.TokenDigest("unknown"), time.Now().Add(time.Hour)); err != nil || created {
		t.Fatalf("unknown account: created=%v err=%v", created, err)
	}
	old := createResetLink(t, ctx, s, " FIRST@EXAMPLE.TEST ", "first-link")
	other := createResetLink(t, ctx, s, "second@example.test", "other-link")
	var originalCreated, originalExpires time.Time
	if err := s.pool.QueryRow(ctx, `SELECT created_at,expires_at FROM password_resets WHERE token=$1`, old).Scan(&originalCreated, &originalExpires); err != nil {
		t.Fatal(err)
	}
	if created, err := s.CreatePasswordReset(ctx, "first@example.test", auth.TokenDigest("throttled"), time.Now().Add(2*time.Hour)); err != nil || created {
		t.Fatalf("cooldown: created=%v err=%v", created, err)
	}
	var unchanged bool
	if err := s.pool.QueryRow(ctx, `SELECT token=$1 AND created_at=$2 AND expires_at=$3 FROM password_resets WHERE user_id=1`, old, originalCreated, originalExpires).Scan(&unchanged); err != nil || !unchanged {
		t.Fatalf("cooldown changed existing link: unchanged=%v err=%v", unchanged, err)
	}
	mustExec(t, ctx, s, `UPDATE password_resets SET created_at=clock_timestamp()-interval '61 seconds' WHERE user_id=1`)
	replacement := createResetLink(t, ctx, s, "first@example.test", "replacement-link")
	for _, token := range []string{old, "replacement-link"} {
		if _, err := s.ResetPassword(ctx, token, "must-not-store"); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatalf("old/raw token accepted: %v", err)
		}
	}
	var beforeReset bool
	if err := s.pool.QueryRow(ctx, `SELECT password='unused' AND character_id=1 AND (SELECT count(*) FROM sessions WHERE user_id=1)=3 FROM users WHERE id=1`).Scan(&beforeReset); err != nil || !beforeReset {
		t.Fatalf("request changed account before consumption: %v %v", beforeReset, err)
	}
	if email, err := s.ResetPassword(ctx, replacement, "new-password-hash"); err != nil || email != "first@example.test" {
		t.Fatalf("reset: email=%q err=%v", email, err)
	}
	var ownOK, otherOK bool
	if err := s.pool.QueryRow(ctx, `SELECT password='new-password-hash' AND character_id IS NULL AND email_verified_at=$1
		AND NOT EXISTS(SELECT 1 FROM sessions WHERE user_id=1) AND NOT EXISTS(SELECT 1 FROM password_resets WHERE user_id=1)
		AND (SELECT count(*) FROM characters WHERE user_id=1)=2 FROM users WHERE id=1`, verified).Scan(&ownOK); err != nil || !ownOK {
		t.Fatalf("reset account postconditions: %v %v", ownOK, err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT password='unused' AND character_id=3 AND email_verified_at IS NOT NULL
		AND (SELECT count(*) FROM sessions WHERE user_id=2)=1 AND EXISTS(SELECT 1 FROM password_resets WHERE token=$1)
		AND (SELECT count(*) FROM characters WHERE user_id=2)=1 FROM users WHERE id=2`, other).Scan(&otherOK); err != nil || !otherOK {
		t.Fatalf("reset changed another account: %v %v", otherOK, err)
	}
	if _, err := s.ResetPassword(ctx, replacement, "reused"); !errors.Is(err, auth.ErrInvalidResetToken) {
		t.Fatalf("reused link: %v", err)
	}
}

func TestPasswordResetInvalidExpiredAndUnverified(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE users SET email_verified_at=NULL WHERE id=1`)
	hash := createResetLink(t, ctx, s, "first@example.test", "unverified-link")
	for _, token := range []string{"", auth.TokenDigest("missing")} {
		if _, err := s.ResetPassword(ctx, token, "rejected"); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatalf("invalid link: %v", err)
		}
	}
	mustExec(t, ctx, s, `UPDATE password_resets SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, hash)
	if _, err := s.ResetPassword(ctx, hash, "rejected"); !errors.Is(err, auth.ErrInvalidResetToken) {
		t.Fatalf("expired link: %v", err)
	}
	var unchanged bool
	if err := s.pool.QueryRow(ctx, `SELECT password='unused' AND email_verified_at IS NULL AND EXISTS(SELECT 1 FROM password_resets WHERE token=$1) FROM users WHERE id=1`, hash).Scan(&unchanged); err != nil || !unchanged {
		t.Fatalf("rejected link changed account: %v %v", unchanged, err)
	}
	mustExec(t, ctx, s, `UPDATE password_resets SET expires_at=clock_timestamp()+interval '1 hour' WHERE token=$1`, hash)
	if _, err := s.ResetPassword(ctx, hash, "unverified-new-hash"); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT password='unverified-new-hash' AND email_verified_at IS NULL FROM users WHERE id=1`).Scan(&unchanged); err != nil || !unchanged {
		t.Fatalf("reset changed verification status: %v %v", unchanged, err)
	}
}

func TestPasswordResetRollsBackWhenRevocationFails(t *testing.T) {
	for _, table := range []string{"sessions", "password_resets"} {
		t.Run(table, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			if err := createStoredSession(t, s, ctx, 1, auth.TokenDigest("live"), time.Now().Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			hash := createResetLink(t, ctx, s, "first@example.test", "rollback")
			mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1;
				CREATE FUNCTION reject_reset_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected delete failure'; END $$;
				CREATE TRIGGER reject_reset_delete BEFORE DELETE ON `+table+` FOR EACH ROW EXECUTE FUNCTION reject_reset_delete()`)
			if _, err := s.ResetPassword(ctx, hash, "must-rollback"); err == nil {
				t.Fatal("expected injected failure")
			}
			var intact bool
			if err := s.pool.QueryRow(ctx, `SELECT password='unused' AND character_id=1 AND email_verified_at IS NOT NULL
				AND (SELECT count(*) FROM sessions WHERE user_id=1)=1 AND EXISTS(SELECT 1 FROM password_resets WHERE token=$1) FROM users WHERE id=1`, hash).Scan(&intact); err != nil || !intact {
				t.Fatalf("failed reset was not atomic: %v %v", intact, err)
			}
			mustExec(t, ctx, s, `DROP TRIGGER reject_reset_delete ON `+table)
			if _, err := s.ResetPassword(ctx, hash, "retry-succeeds"); err != nil {
				t.Fatalf("rollback consumed link: %v", err)
			}
		})
	}
}

func TestPasswordResetConcurrentConsumption(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	hash := createResetLink(t, ctx, s, "first@example.test", "concurrent")
	start := make(chan struct{})
	type result struct {
		password string
		err      error
	}
	done := make(chan result, 2)
	for _, password := range []string{"winner-a", "winner-b"} {
		go func() { <-start; _, err := s.ResetPassword(ctx, hash, password); done <- result{password, err} }()
	}
	close(start)
	winner, successes, invalid := "", 0, 0
	for range 2 {
		r := <-done
		if r.err == nil {
			winner = r.password
			successes++
		} else if errors.Is(r.err, auth.ErrInvalidResetToken) {
			invalid++
		} else {
			t.Fatal(r.err)
		}
	}
	if successes != 1 || invalid != 1 {
		t.Fatalf("success=%d invalid=%d", successes, invalid)
	}
	var stored string
	if err := s.pool.QueryRow(ctx, `SELECT password FROM users WHERE id=1`).Scan(&stored); err != nil || stored != winner {
		t.Fatalf("password=%q winner=%q err=%v", stored, winner, err)
	}
}

func lockResetUser(t *testing.T, ctx context.Context, s *testStorage) (pgx.Tx, int) {
	t.Helper()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	var pid int
	if err := tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	return tx, pid
}

func awaitResetUserLock(t *testing.T, ctx context.Context, s *testStorage, pid int, done <-chan error) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("operation skipped account lock: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
}

func TestPasswordResetRejectsLoginValidatedBeforePasswordChange(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	tx, pid := lockResetUser(t, ctx, s)
	done := make(chan error, 1)
	go func() {
		done <- s.CreateSession(ctx, auth.SessionInput{UserID: 1, PasswordHash: "unused", TokenHash: auth.TokenDigest("stale-login"), ExpiresAt: time.Now().Add(time.Hour)})
	}()
	awaitResetUserLock(t, ctx, s, pid, done)
	if _, err := tx.Exec(ctx, `UPDATE users SET password='changed-under-lock' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("stale login: %v", err)
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE user_id=1`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("stale sessions=%d err=%v", count, err)
	}
}

func TestPasswordResetRechecksLinkAfterAccountLockWait(t *testing.T) {
	for _, scenario := range []string{"expires", "replaced"} {
		t.Run(scenario, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			hash := createResetLink(t, ctx, s, "first@example.test", "waiting-link")
			if scenario == "expires" {
				mustExec(t, ctx, s, `UPDATE password_resets SET expires_at=clock_timestamp()+interval '500 milliseconds' WHERE token=$1`, hash)
			}
			tx, pid := lockResetUser(t, ctx, s)
			done := make(chan error, 1)
			go func() { _, err := s.ResetPassword(ctx, hash, "must-not-store"); done <- err }()
			awaitResetUserLock(t, ctx, s, pid, done)
			if scenario == "replaced" {
				if _, err := tx.Exec(ctx, `UPDATE password_resets SET token=$1 WHERE user_id=1`, auth.TokenDigest("newer-link")); err != nil {
					t.Fatal(err)
				}
			} else {
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for {
					var expired bool
					if err := s.pool.QueryRow(ctx, `SELECT expires_at<=clock_timestamp() FROM password_resets WHERE token=$1`, hash).Scan(&expired); err != nil {
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
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, auth.ErrInvalidResetToken) {
				t.Fatalf("stale link: %v", err)
			}
			var unchanged bool
			if err := s.pool.QueryRow(ctx, `SELECT password='unused' AND (SELECT count(*) FROM password_resets WHERE user_id=1)=1 FROM users WHERE id=1`).Scan(&unchanged); err != nil || !unchanged {
				t.Fatalf("stale reset changed account: %v %v", unchanged, err)
			}
		})
	}
}
