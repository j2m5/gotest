package db_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gotest/internal/maintenance"
)

func cleanupFixture(t *testing.T) (*testStorage, context.Context) {
	t.Helper()
	s, ctx := authTestStorage(t, true)
	executeMigration(t, ctx, s, "20261025000500_auth_expiry_indexes.up.sql")
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password) VALUES('cleanup@example.test','cleanup','unused')`)
	return s, ctx
}

func insertExpired(t *testing.T, ctx context.Context, s *testStorage, count int) {
	t.Helper()
	mustExec(t, ctx, s, `INSERT INTO sessions(user_id,token,expires_at)
		SELECT 1,'session-'||i,clock_timestamp()-interval '1 day' FROM generate_series(1,$1::int) i`, count)
	mustExec(t, ctx, s, `INSERT INTO email_verifications(user_id,token,expired_at)
		SELECT 1,'verification-'||i,clock_timestamp()-interval '1 day' FROM generate_series(1,$1::int) i`, count)
}

func cleanupCounts(t *testing.T, ctx context.Context, s *testStorage, sessions, verifications int) {
	t.Helper()
	var a, b int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM sessions),(SELECT count(*) FROM email_verifications)`).Scan(&a, &b); err != nil {
		t.Fatal(err)
	}
	if a != sessions || b != verifications {
		t.Fatalf("remaining sessions=%d verifications=%d; want %d/%d", a, b, sessions, verifications)
	}
}

func TestCleanupDeletesExpiredRowsInBatchesAndPreservesAccount(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
	var updatedAt time.Time
	if err := s.pool.QueryRow(ctx, `SELECT updated_at FROM users WHERE id=1`).Scan(&updatedAt); err != nil {
		t.Fatal(err)
	}
	insertExpired(t, ctx, s, 5)
	mustExec(t, ctx, s, `INSERT INTO sessions(user_id,token,expires_at) VALUES(1,'live-session',clock_timestamp()+interval '1 day');
		INSERT INTO email_verifications(user_id,token,expired_at) VALUES(1,'live-verification',clock_timestamp()+interval '1 day')`)
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Sessions != 5 || result.Verifications != 5 || result.Batches != 6 || result.Cutoff.IsZero() {
		t.Fatalf("wrong cleanup result: %+v", result)
	}
	cleanupCounts(t, ctx, s, 1, 1)
	var untouched bool
	if err := s.pool.QueryRow(ctx, `SELECT character_id=1 AND updated_at=$1 FROM users WHERE id=1`, updatedAt).Scan(&untouched); err != nil || !untouched {
		t.Fatalf("cleanup changed the account selection: %v", err)
	}
	var live bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sessions WHERE token='live-session') AND EXISTS(SELECT 1 FROM email_verifications WHERE token='live-verification')`).Scan(&live); err != nil || !live {
		t.Fatalf("live credentials removed: %v", err)
	}
	again, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2})
	if err != nil || again.Sessions != 0 || again.Verifications != 0 {
		t.Fatalf("repeat run: %+v %v", again, err)
	}
}

func TestCleanupPreviewDoesNotDelete(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 3)
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2, DryRun: true})
	if err != nil || result.Sessions != 3 || result.Verifications != 3 || result.Batches != 0 {
		t.Fatalf("preview: %+v %v", result, err)
	}
	cleanupCounts(t, ctx, s, 3, 3)
}

func TestCleanupSkipsBusyCredentialsAndProcessesThemOnNextRun(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 3)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT id FROM sessions WHERE token='session-1' FOR UPDATE;
		SELECT id FROM email_verifications WHERE token='verification-1' FOR UPDATE`); err != nil {
		t.Fatal(err)
	}
	short, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	result, err := maintenance.Run(short, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || result.Sessions != 2 || result.Verifications != 2 {
		t.Fatalf("cleanup waited for or removed locked rows: %+v %v", result, err)
	}
	cleanupCounts(t, ctx, s, 1, 1)
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	result, err = maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || result.Sessions != 1 || result.Verifications != 1 {
		t.Fatalf("unlocked rows not cleaned: %+v %v", result, err)
	}
}

func TestCleanupRejectsInvalidOptionsAndCancelledContextWithoutWrites(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 1)
	for _, size := range []int{-1, 0, 10001} {
		if _, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: size}); err == nil {
			t.Fatalf("invalid batch size %d accepted", size)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := maintenance.Run(cancelled, s.pool, maintenance.Options{BatchSize: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	cleanupCounts(t, ctx, s, 1, 1)
}

func TestCleanupKeepsEarlierBatchesAfterStatementFailure(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 5)
	mustExec(t, ctx, s, `CREATE FUNCTION reject_cleanup_row() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF OLD.token='session-3' THEN RAISE EXCEPTION 'private-trigger-detail'; END IF; RETURN OLD; END $$;
		CREATE TRIGGER reject_cleanup BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION reject_cleanup_row()`)
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2})
	if err == nil || result.Sessions != 2 || result.Verifications != 2 || result.Batches != 2 {
		t.Fatalf("partial progress misreported: %+v %v", result, err)
	}
	cleanupCounts(t, ctx, s, 3, 3)
	mustExec(t, ctx, s, `DROP TRIGGER reject_cleanup ON sessions; DROP FUNCTION reject_cleanup_row()`)
	result, err = maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2})
	if err != nil || result.Sessions != 3 || result.Verifications != 3 {
		t.Fatalf("retry did not finish remaining rows: %+v %v", result, err)
	}
}

func TestCleanupCancelsAnInFlightBatchAndKeepsConfirmedProgress(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 5)
	mustExec(t, ctx, s, `CREATE FUNCTION delay_cleanup_row() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF OLD.token='session-3' THEN PERFORM pg_sleep(10); END IF; RETURN OLD; END $$;
		CREATE TRIGGER delay_cleanup BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION delay_cleanup_row()`)
	deadline, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	result, err := maintenance.Run(deadline, s.pool, maintenance.Options{BatchSize: 2})
	if !errors.Is(err, context.DeadlineExceeded) || result.Sessions != 2 || result.Verifications != 2 {
		t.Fatalf("timeout lost partial progress: %+v %v", result, err)
	}
	cleanupCounts(t, ctx, s, 3, 3)
}

func TestCleanupUsesOneCutoffForTheWholePass(t *testing.T) {
	s, ctx := cleanupFixture(t)
	mustExec(t, ctx, s, `INSERT INTO sessions(user_id,token,expires_at) VALUES
		(1,'old',clock_timestamp()-interval '1 day'),(1,'later',clock_timestamp()+interval '1 day');
		CREATE FUNCTION expire_during_cleanup() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF OLD.token='old' THEN
			PERFORM pg_sleep(0.01); UPDATE sessions SET expires_at=clock_timestamp() WHERE token='later';
		END IF; RETURN OLD; END $$;
		CREATE TRIGGER expire_during_cleanup BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION expire_during_cleanup()`)
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || result.Sessions != 1 {
		t.Fatalf("cutoff moved during cleanup: %+v %v", result, err)
	}
	cleanupCounts(t, ctx, s, 1, 0)
	var newer bool
	if err := s.pool.QueryRow(ctx, `SELECT expires_at>$1 AND expires_at<=clock_timestamp() FROM sessions WHERE token='later'`, result.Cutoff).Scan(&newer); err != nil || !newer {
		t.Fatalf("fixture did not expire during the pass: %v", err)
	}
	result, err = maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil || result.Sessions != 1 {
		t.Fatalf("next pass did not advance cutoff: %+v %v", result, err)
	}
}

func TestConcurrentCleanupsDoNotDoubleCountOrLoseRows(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 20)
	var wg sync.WaitGroup
	results := make([]maintenance.Result, 2)
	errors := make([]error, 2)
	for i := range 2 {
		wg.Go(func() { results[i], errors[i] = maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 2}) })
	}
	wg.Wait()
	if errors[0] != nil || errors[1] != nil || results[0].Sessions+results[1].Sessions != 20 || results[0].Verifications+results[1].Verifications != 20 {
		t.Fatalf("concurrent cleanup results: %+v errors=%v", results, errors)
	}
	cleanupCounts(t, ctx, s, 0, 0)
}

func TestExpiryIndexMigrationRoundTripPreservesCredentials(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 2)
	executeMigration(t, ctx, s, "20261025000500_auth_expiry_indexes.down.sql")
	cleanupCounts(t, ctx, s, 2, 2)
	executeMigration(t, ctx, s, "20261025000500_auth_expiry_indexes.up.sql")
	cleanupCounts(t, ctx, s, 2, 2)
	var indexes int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema() AND
		((tablename='sessions' AND indexdef LIKE '%(expires_at, id)') OR
		 (tablename='email_verifications' AND indexdef LIKE '%(expired_at, id)'))`).Scan(&indexes); err != nil || indexes != 2 {
		t.Fatalf("expiry indexes missing after reapply: %d %v", indexes, err)
	}
}
