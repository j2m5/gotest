package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/maintenance"
)

func TestQueueStatusMixedStatesAreReadOnly(t *testing.T) {
	s, ctx, q := queueStorage(t)
	for n := 1; n <= 9; n++ {
		if err := q.EnqueueResetRequest(ctx, queueMessage(n)); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET created_at=clock_timestamp()-interval '2 hours', available_at=clock_timestamp()-interval '1 hour';
		UPDATE mail_jobs SET available_at=clock_timestamp()+interval '1 hour' WHERE id=repeat('0',63)||'2';
		UPDATE mail_jobs SET leased_until=clock_timestamp()+interval '1 hour', lease_token=repeat('b',64), attempts=8 WHERE id=repeat('0',63)||'3';
		UPDATE mail_jobs SET attempts=8 WHERE id=repeat('0',63)||'4';
		UPDATE mail_jobs SET failed_at=clock_timestamp()-interval '2 days' WHERE id IN (repeat('0',63)||'5',repeat('0',63)||'6');
		UPDATE mail_jobs SET expires_at=clock_timestamp()-interval '1 hour' WHERE id IN (repeat('0',63)||'6',repeat('0',63)||'7');
		UPDATE mail_jobs SET leased_until=clock_timestamp()-interval '1 hour',lease_token=repeat('b',64) WHERE id=repeat('0',63)||'8';
		UPDATE mail_jobs SET created_at=clock_timestamp()+interval '1 hour' WHERE id=repeat('0',63)||'9'`)
	var before string
	if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(m) ORDER BY id)::text FROM mail_jobs m`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	var start, end time.Time
	if err := s.pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&start); err != nil {
		t.Fatal(err)
	}
	result, err := maintenance.QueueStatus(ctx, s.pool)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&end); err != nil {
		t.Fatal(err)
	}
	if result.Total != 9 || result.Pending != 4 || result.Due != 3 || result.Leased != 1 || result.Failed != 2 || result.Expired != 2 {
		t.Fatalf("wrong status: %+v", result)
	}
	if result.Cutoff.Before(start) || result.Cutoff.After(end) || result.Cutoff.Location() != time.UTC {
		t.Fatal("cutoff is not the database clock in UTC")
	}
	if result.OldestPendingAgeSeconds == nil || *result.OldestPendingAgeSeconds < 7200 || *result.OldestPendingAgeSeconds > 7210 {
		t.Fatal("oldest pending age did not use created_at")
	}
	var after string
	if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(m) ORDER BY id)::text FROM mail_jobs m`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("queue status mutated jobs")
	}
}

func TestQueueStatusEmpty(t *testing.T) {
	s, ctx, _ := queueStorage(t)
	r, err := maintenance.QueueStatus(ctx, s.pool)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 0 || r.Pending != 0 || r.Due != 0 || r.Leased != 0 || r.Failed != 0 || r.Expired != 0 || r.OldestPendingAgeSeconds != nil || r.Cutoff.IsZero() {
		t.Fatalf("empty status: %+v", r)
	}
}

func TestQueueStatusExactClockBoundariesAndFutureCreation(t *testing.T) {
	s, ctx, _ := queueStorage(t)
	// Match the existing exact-boundary test pattern with an isolated DB clock.
	mustExec(t, ctx, s, `CREATE FUNCTION clock_timestamp() RETURNS timestamptz LANGUAGE SQL VOLATILE AS $$ SELECT '2040-01-01 12:00:00+00'::timestamptz $$`)
	config := s.pool.Config()
	config.ConnConfig.RuntimeParams["search_path"] += ",pg_catalog"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	s = &testStorage{pool: pool}
	mustExec(t, ctx, s, `ALTER TABLE mail_jobs RENAME TO mail_jobs_fixture;
		CREATE VIEW mail_jobs AS SELECT clock_timestamp()+interval '1 hour' AS created_at,
		clock_timestamp()+expiry AS expires_at, clock_timestamp() AS available_at,
		CASE WHEN lease IS NULL THEN NULL ELSE clock_timestamp()+lease END AS leased_until,
		NULL::timestamptz AS failed_at, attempts FROM (VALUES
		(interval '0 seconds',NULL::interval,0),
		(interval '1 hour',interval '0 seconds',0),
		(interval '1 hour',interval '0 seconds',8),
		(interval '1 hour',interval '1 hour',8)) f(expiry,lease,attempts)`)
	r, err := maintenance.QueueStatus(ctx, s.pool)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 4 || r.Expired != 1 || r.Pending != 1 || r.Due != 1 || r.Failed != 1 || r.Leased != 1 || r.OldestPendingAgeSeconds == nil || *r.OldestPendingAgeSeconds != 0 {
		t.Fatalf("exact boundaries: %+v", r)
	}
}
