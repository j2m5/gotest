package db_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"gotest/internal/db"
	"gotest/internal/notification"
)

func queueStorage(t *testing.T) (*testStorage, context.Context, *db.MailQueue) {
	t.Helper()
	s, ctx := authTestStorage(t, true)
	executeMigration(t, ctx, s, "20261027000200_mail_jobs.up.sql")
	return s, ctx, db.NewMailQueue(s.pool)
}

func queueMessage(n int) notification.Message {
	return notification.Message{ID: fmt.Sprintf("%064x", n), Kind: notification.ResetRequest,
		KeyID: strings.Repeat("a", 64), Ciphertext: []byte("encrypted test payload"),
		ExpiresAt: time.Now().UTC().Truncate(time.Microsecond).Add(time.Hour)}
}

func queueClaim(t *testing.T, ctx context.Context, q *db.MailQueue) notification.Job {
	t.Helper()
	j, err := q.Claim(ctx)
	if err != nil || j == nil {
		t.Fatalf("claim: %+v %v", j, err)
	}
	return *j
}

func TestMailQueueMigrationUpDownUp(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password) VALUES('queue@example.test','queue','unchanged')`)
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, "20261027000200_mail_jobs."+direction+".sql")
		s.pool.Reset()
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT to_regclass(current_schema()||'.mail_jobs') IS NOT NULL`).Scan(&exists); err != nil || exists != (direction == "up") {
			t.Fatalf("%s table exists=%v err=%v", direction, exists, err)
		}
		var intact bool
		if err := s.pool.QueryRow(ctx, `SELECT count(*)=1 AND bool_and(password='unchanged') FROM users`).Scan(&intact); err != nil || !intact {
			t.Fatalf("migration changed existing accounts: %v %v", intact, err)
		}
		if direction == "down" {
			continue
		}
		var schemaOK bool
		if err := s.pool.QueryRow(ctx, `SELECT
			(SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='mail_jobs' AND column_name IN ('created_at','expires_at','available_at','leased_until','failed_at') AND data_type='timestamp with time zone')=5
			AND (SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema() AND tablename='mail_jobs')=4`).Scan(&schemaOK); err != nil || !schemaOK {
			t.Fatalf("mail queue schema: %v %v", schemaOK, err)
		}
		mustExec(t, ctx, s, `INSERT INTO mail_jobs(id,kind,key_id,payload,user_id,expires_at)
			VALUES(repeat('1',64),'verification',repeat('a',64),'ciphertext',1,clock_timestamp()+interval '1 hour')`)
	}
	for _, change := range []string{
		`id='invalid'`, `key_id=repeat('A',64)`, `kind='unknown'`, `user_id=NULL`,
		`kind='reset_request'`, `user_id=99999`, `payload=''`, `attempts=-1`, `attempts=9`,
		`lease_token=repeat('b',64)`, `leased_until=clock_timestamp()`,
		`lease_token=repeat('b',64),leased_until=clock_timestamp(),failed_at=clock_timestamp()`,
	} {
		if _, err := s.pool.Exec(ctx, `UPDATE mail_jobs SET `+change); err == nil {
			t.Fatalf("invalid state accepted: %s", change)
		}
	}
	mustExec(t, ctx, s, `DELETE FROM users WHERE id=1`)
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM mail_jobs`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("account cascade: count=%d err=%v", count, err)
	}
}

func TestMailQueueSkipsLockedAndUnavailableJobs(t *testing.T) {
	s, ctx, q := queueStorage(t)
	for n := 1; n <= 4; n++ {
		if err := q.EnqueueResetRequest(ctx, queueMessage(n)); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET available_at=clock_timestamp()+interval '1 hour' WHERE id=$1`, queueMessage(2).ID)
	mustExec(t, ctx, s, `UPDATE mail_jobs SET attempts=8 WHERE id=$1`, queueMessage(3).ID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT id FROM mail_jobs WHERE id=$1 FOR UPDATE`, queueMessage(1).ID); err != nil {
		t.Fatal(err)
	}
	claimCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	j := queueClaim(t, claimCtx, q)
	if j.ID != queueMessage(4).ID {
		t.Fatalf("claimed locked/unavailable job: %+v", j)
	}
	var leaseSeconds float64
	if err := s.pool.QueryRow(ctx, `SELECT extract(epoch FROM leased_until-clock_timestamp()) FROM mail_jobs WHERE id=$1`, j.ID).Scan(&leaseSeconds); err != nil || leaseSeconds < 115 || leaseSeconds > 120 {
		t.Fatalf("lease duration: %f %v", leaseSeconds, err)
	}
	if j, err := q.Claim(claimCtx); err != nil || j != nil {
		t.Fatalf("unavailable claim: %+v %v", j, err)
	}
}

func TestMailQueuePersistsAndClaimsOnceConcurrently(t *testing.T) {
	s, ctx, q := queueStorage(t)
	for n := 1; n <= 16; n++ {
		if err := q.EnqueueResetRequest(ctx, queueMessage(n)); err != nil {
			t.Fatal(err)
		}
	}
	// A newly constructed store reads the same durable jobs.
	q = db.NewMailQueue(s.pool)
	var wg sync.WaitGroup
	results := make(chan *notification.Job, 24)
	errs := make(chan error, 24)
	for range 24 {
		wg.Go(func() { j, err := q.Claim(ctx); results <- j; errs <- err })
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	for j := range results {
		if j == nil {
			continue
		}
		if seen[j.ID] || j.Attempts != 1 || len(j.Lease) != 64 || j.CreatedAt.IsZero() {
			t.Fatalf("invalid or duplicate claim: %+v", j)
		}
		seen[j.ID] = true
	}
	if len(seen) != 16 {
		t.Fatalf("claimed %d jobs", len(seen))
	}
}

func TestMailQueueReclaimRejectsStaleLease(t *testing.T) {
	s, ctx, q := queueStorage(t)
	if err := q.EnqueueResetRequest(ctx, queueMessage(1)); err != nil {
		t.Fatal(err)
	}
	old := queueClaim(t, ctx, q)
	if owns, err := q.Owns(ctx, old); err != nil || !owns {
		t.Fatalf("owns=%v err=%v", owns, err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET leased_until=clock_timestamp()-interval '1 second'`)
	if owns, err := q.Owns(ctx, old); err != nil || owns {
		t.Fatalf("expired owns=%v err=%v", owns, err)
	}
	if err := q.Complete(ctx, old); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("expired complete: %v", err)
	}
	if err := q.Retry(ctx, old, 0, false); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("expired retry: %v", err)
	}
	current := queueClaim(t, ctx, q)
	if current.Lease == old.Lease || current.Attempts != 2 {
		t.Fatalf("reclaim: %+v", current)
	}
	if err := q.Complete(ctx, old); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("stale complete: %v", err)
	}
	if err := q.Retry(ctx, old, 0, true); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("stale retry: %v", err)
	}
	if err := q.Complete(ctx, current); err != nil {
		t.Fatal(err)
	}
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("completed job reclaimed: %+v %v", j, err)
	}
}

func TestMailQueueRetryDueExpiryAndAttemptLimit(t *testing.T) {
	s, ctx, q := queueStorage(t)
	if err := q.EnqueueResetRequest(ctx, queueMessage(1)); err != nil {
		t.Fatal(err)
	}
	j := queueClaim(t, ctx, q)
	if err := q.Retry(ctx, j, time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("backoff ignored: %+v %v", j, err)
	}
	var scheduled bool
	if err := s.pool.QueryRow(ctx, `SELECT available_at > clock_timestamp()+interval '50 seconds' AND lease_token IS NULL AND leased_until IS NULL FROM mail_jobs`).Scan(&scheduled); err != nil || !scheduled {
		t.Fatalf("retry schedule: %v %v", scheduled, err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET available_at=clock_timestamp()-interval '1 second', attempts=7`)
	j = queueClaim(t, ctx, q)
	if j.Attempts != 8 {
		t.Fatalf("attempts=%d", j.Attempts)
	}
	if err := q.Retry(ctx, j, 0, false); err != nil {
		t.Fatal(err)
	}
	var failed bool
	if err := s.pool.QueryRow(ctx, `SELECT failed_at IS NOT NULL AND payload IS NOT NULL AND lease_token IS NULL FROM mail_jobs`).Scan(&failed); err != nil || !failed {
		t.Fatalf("failed payload retention: %v %v", failed, err)
	}
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("exhausted claim: %+v %v", j, err)
	}
	if err := q.EnqueueResetRequest(ctx, queueMessage(2)); err != nil {
		t.Fatal(err)
	}
	j = queueClaim(t, ctx, q)
	mustExec(t, ctx, s, `UPDATE mail_jobs SET expires_at=clock_timestamp()+interval '2 seconds' WHERE id=$1`, j.ID)
	if err := q.Retry(ctx, j, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT available_at=expires_at FROM mail_jobs WHERE id=$1`, j.ID).Scan(&scheduled); err != nil || !scheduled {
		t.Fatalf("retry crossed expiry: %v %v", scheduled, err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET expires_at=clock_timestamp()-interval '1 second', available_at=clock_timestamp()-interval '2 seconds'`)
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("expired claim: %+v %v", j, err)
	}
}

func TestMailQueueCapacityAndKeyMismatch(t *testing.T) {
	s, ctx, q := queueStorage(t)
	// No account is created. Intake must not depend on account existence.
	mustExec(t, ctx, s, `INSERT INTO mail_jobs(id,kind,key_id,payload,expires_at)
		SELECT lpad(to_hex(n),64,'0'),'reset_request',repeat('a',64),'ciphertext',clock_timestamp()+interval '1 hour' FROM generate_series(1,1023) n`)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, n := range []int{1024, 1025} {
		wg.Go(func() { errs <- q.EnqueueResetRequest(ctx, queueMessage(n)) })
	}
	wg.Wait()
	close(errs)
	ok, full := 0, 0
	for err := range errs {
		if err == nil {
			ok++
		} else if errors.Is(err, notification.ErrQueueFull) {
			full++
		} else {
			t.Fatal(err)
		}
	}
	if ok != 1 || full != 1 {
		t.Fatalf("capacity race: ok=%d full=%d", ok, full)
	}
	if err := q.ValidateKey(ctx, strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	if err := q.ValidateKey(ctx, strings.Repeat("b", 64)); err == nil {
		t.Fatal("wrong key accepted")
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET failed_at=clock_timestamp()`)
	if err := q.ValidateKey(ctx, strings.Repeat("b", 64)); err == nil {
		t.Fatal("wrong key accepted for retained failures")
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET expires_at=clock_timestamp()-interval '1 second'`)
	if err := q.ValidateKey(ctx, strings.Repeat("b", 64)); err != nil {
		t.Fatal(err)
	}
	if err := q.EnqueueResetRequest(ctx, queueMessage(1026)); err != nil {
		t.Fatalf("expired jobs occupy capacity: %v", err)
	}
}

func TestMailQueueRejectsMalformedIncomingAndPermanentRetry(t *testing.T) {
	s, ctx, q := queueStorage(t)
	for _, mutate := range []func(*notification.Message){
		func(m *notification.Message) { m.Kind = notification.Verification },
		func(m *notification.Message) { m.ID = "raw-token" },
		func(m *notification.Message) { m.KeyID = strings.Repeat("A", 64) },
		func(m *notification.Message) { m.Ciphertext = nil },
		func(m *notification.Message) { m.ExpiresAt = time.Now().Add(-time.Second) },
	} {
		m := queueMessage(1)
		mutate(&m)
		if err := q.EnqueueResetRequest(ctx, m); err == nil {
			t.Fatal("malformed incoming accepted")
		}
	}
	if err := q.EnqueueResetRequest(ctx, queueMessage(1)); err != nil {
		t.Fatal(err)
	}
	j := queueClaim(t, ctx, q)
	if err := q.Retry(ctx, j, 0, true); err != nil {
		t.Fatal(err)
	}
	var retained bool
	if err := s.pool.QueryRow(ctx, `SELECT failed_at IS NOT NULL AND payload=$1 AND lease_token IS NULL FROM mail_jobs`, j.Ciphertext).Scan(&retained); err != nil || !retained {
		t.Fatalf("permanent failure: %v %v", retained, err)
	}
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("permanent claim: %+v %v", j, err)
	}
}

func TestMailQueueRetiresAbandonedFinalAttemptAfterLeaseExpires(t *testing.T) {
	s, ctx, q := queueStorage(t)
	if err := q.EnqueueResetRequest(ctx, queueMessage(1)); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET attempts=7`)
	mustExec(t, ctx, s, `INSERT INTO mail_jobs(id,kind,key_id,payload,expires_at,available_at)
		SELECT lpad(to_hex(n),64,'0'),'reset_request',repeat('a',64),'ciphertext',
		clock_timestamp()+interval '1 hour',clock_timestamp()+interval '30 minutes' FROM generate_series(2,1024) n`)
	final := queueClaim(t, ctx, q)
	if final.Attempts != 8 {
		t.Fatalf("final attempts=%d", final.Attempts)
	}
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("active final lease reclaimed: %+v %v", j, err)
	}
	var active bool
	if err := s.pool.QueryRow(ctx, `SELECT failed_at IS NULL AND lease_token=$2 FROM mail_jobs WHERE id=$1`, final.ID, final.Lease).Scan(&active); err != nil || !active {
		t.Fatalf("active final lease retired: %v %v", active, err)
	}
	if owns, err := q.Owns(ctx, final); err != nil || !owns {
		t.Fatalf("final lease owns=%v err=%v", owns, err)
	}
	if err := q.EnqueueResetRequest(ctx, queueMessage(1025)); !errors.Is(err, notification.ErrQueueFull) {
		t.Fatalf("active final attempt omitted from capacity: %v", err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET leased_until=clock_timestamp()-interval '1 second' WHERE id=$1`, final.ID)
	if j, err := q.Claim(ctx); err != nil || j != nil {
		t.Fatalf("exhausted lease reclaimed: %+v %v", j, err)
	}
	var retired bool
	if err := s.pool.QueryRow(ctx, `SELECT failed_at IS NOT NULL AND lease_token IS NULL AND leased_until IS NULL AND payload=$2 AND attempts=8 FROM mail_jobs WHERE id=$1`, final.ID, final.Ciphertext).Scan(&retired); err != nil || !retired {
		t.Fatalf("abandoned final attempt not retained as failed: %v %v", retired, err)
	}
	if err := q.EnqueueResetRequest(ctx, queueMessage(1025)); err != nil {
		t.Fatalf("failed final attempt still consumes capacity: %v", err)
	}
	if err := q.Complete(ctx, final); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("retired complete: %v", err)
	}
	if err := q.Retry(ctx, final, 0, false); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("retired retry: %v", err)
	}
}
