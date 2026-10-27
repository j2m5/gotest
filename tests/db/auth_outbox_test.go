package db_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/mailer"
	"gotest/internal/maintenance"
	"gotest/internal/notification"
)

func outboxMessage(t *testing.T, ctx context.Context, s *testStorage, kind notification.Kind) notification.Message {
	t.Helper()
	var m notification.Message
	if err := s.pool.QueryRow(ctx, `SELECT id,kind,key_id,payload,expires_at FROM mail_jobs WHERE kind=$1`, string(kind)).Scan(&m.ID, &m.Kind, &m.KeyID, &m.Ciphertext, &m.ExpiresAt); err != nil {
		t.Fatal(err)
	}
	return m
}

func outboxContent(t *testing.T, message notification.Message) notification.Content {
	t.Helper()
	content, err := testCipher().Open(message)
	if err != nil {
		t.Fatal("cannot decrypt persisted notification")
	}
	return content
}

func rejectOutboxKind(t *testing.T, ctx context.Context, s *testStorage, kind notification.Kind) {
	t.Helper()
	// The kind comes only from fixed enum values in these tests.
	mustExec(t, ctx, s, `CREATE FUNCTION reject_outbox_insert() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF NEW.kind='`+string(kind)+`' THEN RAISE EXCEPTION 'fixture rejects notification'; END IF; RETURN NEW; END $$;
		CREATE TRIGGER reject_outbox_insert BEFORE INSERT ON mail_jobs FOR EACH ROW EXECUTE FUNCTION reject_outbox_insert()`)
}

func allowOutbox(t *testing.T, ctx context.Context, s *testStorage) {
	t.Helper()
	mustExec(t, ctx, s, `DROP TRIGGER reject_outbox_insert ON mail_jobs; DROP FUNCTION reject_outbox_insert()`)
}

func outboxCheck(t *testing.T, ctx context.Context, s *testStorage, sql string, args ...any) {
	t.Helper()
	var ok bool
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&ok); err != nil || !ok {
		t.Fatalf("outbox/account invariant failed: %v", err)
	}
}

func TestAuthRegistrationCommitsEncryptedMail(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	if err := service.Register(ctx, auth.RegisterInput{Email: "queued@example.test", Login: "queued", Password: "test-password"}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM mail_jobs WHERE user_id=(SELECT id FROM users WHERE email='queued@example.test') AND kind='verification'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("registration committed without notification: %d", count)
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT payload FROM mail_jobs WHERE user_id=(SELECT id FROM users WHERE email='queued@example.test')`).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(payload, []byte("queued@example.test")) || strings.Contains(string(payload), "test-password") {
		t.Fatal("private data persisted in plaintext")
	}
	content := outboxContent(t, outboxMessage(t, ctx, s, notification.Verification))
	if content.Email != "queued@example.test" || len(content.Token) != 64 {
		t.Fatal("incorrect encrypted verification content")
	}
	outboxCheck(t, ctx, s, `SELECT token=$1 FROM email_verifications WHERE user_id=(SELECT id FROM users WHERE email='queued@example.test')`, auth.TokenDigest(content.Token))
}

func TestAuthOutboxRegistrationRollback(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	rejectOutboxKind(t, ctx, s, notification.Verification)
	if err := service.Register(ctx, auth.RegisterInput{Email: "rollback-outbox@example.test", Login: "rollback-outbox", Password: "test-password"}); err == nil {
		t.Fatal("mail insert failure accepted")
	}
	outboxCheck(t, ctx, s, `SELECT NOT EXISTS(SELECT 1 FROM users WHERE email='rollback-outbox@example.test') AND NOT EXISTS(SELECT 1 FROM email_verifications) AND NOT EXISTS(SELECT 1 FROM mail_jobs)`)
	allowOutbox(t, ctx, s)
	if err := service.Register(ctx, auth.RegisterInput{Email: "rollback-outbox@example.test", Login: "rollback-outbox", Password: "test-password"}); err != nil {
		t.Fatal(err)
	}
	content := outboxContent(t, outboxMessage(t, ctx, s, notification.Verification))
	outboxCheck(t, ctx, s, `SELECT token=$1 FROM email_verifications`, auth.TokenDigest(content.Token))
}

func TestAuthOutboxResendRollbackReplacementAndConsumption(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	if err := service.Register(ctx, auth.RegisterInput{Email: "resend-outbox@example.test", Login: "resend-outbox", Password: "test-password"}); err != nil {
		t.Fatal(err)
	}
	old := outboxMessage(t, ctx, s, notification.Verification)
	oldContent := outboxContent(t, old)
	rejectOutboxKind(t, ctx, s, notification.Verification)
	if err := service.ResendVerification(ctx, "resend-outbox@example.test", "test-password"); err == nil {
		t.Fatal("resend accepted mail insert failure")
	}
	outboxCheck(t, ctx, s, `SELECT (SELECT count(*) FROM email_verifications WHERE token=$1)=1 AND (SELECT count(*) FROM mail_jobs WHERE id=$2 AND payload=$3)=1`, auth.TokenDigest(oldContent.Token), old.ID, old.Ciphertext)
	allowOutbox(t, ctx, s)
	if err := service.ResendVerification(ctx, "resend-outbox@example.test", "test-password"); err != nil {
		t.Fatal(err)
	}
	current := outboxMessage(t, ctx, s, notification.Verification)
	content := outboxContent(t, current)
	if current.ID == old.ID || content.Token == oldContent.Token {
		t.Fatal("resend reused prior token/job")
	}
	outboxCheck(t, ctx, s, `SELECT (SELECT count(*) FROM mail_jobs)=1 AND (SELECT count(*) FROM email_verifications)=1 AND EXISTS(SELECT 1 FROM email_verifications WHERE token=$1)`, auth.TokenDigest(content.Token))
	if err := service.VerifyEmail(ctx, oldContent.Token); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("superseded token: %v", err)
	}
	if err := service.VerifyEmail(ctx, content.Token); err != nil {
		t.Fatal(err)
	}
	outboxCheck(t, ctx, s, `SELECT NOT EXISTS(SELECT 1 FROM email_verifications) AND NOT EXISTS(SELECT 1 FROM mail_jobs) AND EXISTS(SELECT 1 FROM users WHERE email='resend-outbox@example.test' AND email_verified_at IS NOT NULL)`)
}

func TestAuthOutboxPasswordChangeRollsBackAllAccountChanges(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	hash := createResetLink(t, ctx, s, "first@example.test", "atomic-reset")
	if err := createStoredSession(t, s, ctx, 1, auth.TokenDigest("atomic-session"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
	var verified time.Time
	if err := s.pool.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id=1`).Scan(&verified); err != nil {
		t.Fatal(err)
	}
	prior := outboxMessage(t, ctx, s, notification.PasswordReset)
	rejectOutboxKind(t, ctx, s, notification.PasswordChanged)
	if _, err := s.ResetPassword(ctx, hash, "new-password-hash"); err == nil {
		t.Fatal("reset accepted mail insert failure")
	}
	outboxCheck(t, ctx, s, `SELECT password='unused' AND character_id=1 AND email_verified_at=$1
		AND EXISTS(SELECT 1 FROM sessions WHERE user_id=1) AND EXISTS(SELECT 1 FROM password_resets WHERE token=$2)
		AND EXISTS(SELECT 1 FROM mail_jobs WHERE id=$3 AND payload=$4) FROM users WHERE id=1`, verified, hash, prior.ID, prior.Ciphertext)
	allowOutbox(t, ctx, s)
	if _, err := s.ResetPassword(ctx, hash, "new-password-hash"); err != nil {
		t.Fatal(err)
	}
	outboxCheck(t, ctx, s, `SELECT password='new-password-hash' AND character_id IS NULL AND email_verified_at=$1
		AND NOT EXISTS(SELECT 1 FROM sessions WHERE user_id=1) AND NOT EXISTS(SELECT 1 FROM password_resets WHERE user_id=1)
		AND NOT EXISTS(SELECT 1 FROM mail_jobs WHERE kind='password_reset') AND (SELECT count(*) FROM mail_jobs WHERE kind='password_changed' AND user_id=1)=1 FROM users WHERE id=1`, verified)
	content := outboxContent(t, outboxMessage(t, ctx, s, notification.PasswordChanged))
	if content.Email != "first@example.test" || content.Token != "" {
		t.Fatal("invalid password changed content")
	}
}

func TestAuthOutboxResetInboxAtomicReplayCooldownAndUnknown(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	q := db.NewMailQueue(s.pool)
	for _, email := range []string{"first@example.test", "unknown@example.test"} {
		if err := service.RequestPasswordReset(ctx, email); err != nil {
			t.Fatal(err)
		}
	}
	outboxCheck(t, ctx, s, `SELECT (SELECT count(*) FROM mail_jobs WHERE kind='reset_request' AND user_id IS NULL)=2 AND NOT EXISTS(SELECT 1 FROM password_resets)`)
	known, unknown := queueClaim(t, ctx, q), queueClaim(t, ctx, q)
	knownContent, unknownContent := outboxContent(t, known.Message), outboxContent(t, unknown.Message)
	if knownContent.Email != "first@example.test" || unknownContent.Email != "unknown@example.test" || knownContent.Token != "" || unknownContent.Token != "" {
		t.Fatal("incorrect reset intake payload")
	}
	for _, job := range []notification.Job{known, unknown} {
		if bytes.Contains(job.Ciphertext, []byte("example.test")) {
			t.Fatal("reset intake stored email in plaintext")
		}
	}
	request := auth.QueuedResetRequest{ID: known.ID, Lease: known.Lease, Email: knownContent.Email}
	rejectOutboxKind(t, ctx, s, notification.PasswordReset)
	if err := service.ProcessPasswordReset(ctx, request); err == nil {
		t.Fatal("inbox processing accepted outgoing failure")
	}
	outboxCheck(t, ctx, s, `SELECT EXISTS(SELECT 1 FROM mail_jobs WHERE id=$1 AND lease_token=$2 AND payload=$3) AND NOT EXISTS(SELECT 1 FROM password_resets) AND NOT EXISTS(SELECT 1 FROM mail_jobs WHERE kind='password_reset')`, known.ID, known.Lease, known.Ciphertext)
	allowOutbox(t, ctx, s)
	if err := service.ProcessPasswordReset(ctx, request); err != nil {
		t.Fatal(err)
	}
	outgoing := outboxMessage(t, ctx, s, notification.PasswordReset)
	content := outboxContent(t, outgoing)
	outboxCheck(t, ctx, s, `SELECT NOT EXISTS(SELECT 1 FROM mail_jobs WHERE id=$1) AND EXISTS(SELECT 1 FROM password_resets WHERE user_id=1 AND token=$2)`, known.ID, auth.TokenDigest(content.Token))
	if err := service.ProcessPasswordReset(ctx, request); !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("replayed input: %v", err)
	}
	if err := service.ProcessPasswordReset(ctx, auth.QueuedResetRequest{ID: unknown.ID, Lease: unknown.Lease, Email: unknownContent.Email}); err != nil {
		t.Fatal(err)
	}
	// Keep the outgoing delivery pending while claiming only the next inbox job.
	mustExec(t, ctx, s, `UPDATE mail_jobs SET available_at=clock_timestamp()+interval '10 minutes' WHERE id=$1`, outgoing.ID)
	if err := service.RequestPasswordReset(ctx, "first@example.test"); err != nil {
		t.Fatal(err)
	}
	cooldown := queueClaim(t, ctx, q)
	if err := service.ProcessPasswordReset(ctx, auth.QueuedResetRequest{ID: cooldown.ID, Lease: cooldown.Lease, Email: "first@example.test"}); err != nil {
		t.Fatal(err)
	}
	outboxCheck(t, ctx, s, `SELECT (SELECT count(*) FROM password_resets)=1 AND EXISTS(SELECT 1 FROM password_resets WHERE user_id=1 AND token=$1)
		AND (SELECT count(*) FROM mail_jobs)=1 AND EXISTS(SELECT 1 FROM mail_jobs WHERE id=$2 AND payload=$3)`, auth.TokenDigest(content.Token), outgoing.ID, outgoing.Ciphertext)
}

func TestAuthOutboxExpiredClaimCannotIssuePasswordReset(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	queue := db.NewMailQueue(s.pool)
	if err := service.RequestPasswordReset(ctx, "first@example.test"); err != nil {
		t.Fatal(err)
	}
	job := queueClaim(t, ctx, queue)
	owned, err := queue.Owns(ctx, job)
	if err != nil || !owned {
		t.Fatalf("new claim not owned: %v", err)
	}
	content := outboxContent(t, job.Message)
	// The worker has already decrypted the request. Advance only its SQL expiry
	// to model a delay before the auth transaction, leaving the lease valid.
	mustExec(t, ctx, s, `UPDATE mail_jobs SET expires_at=clock_timestamp()-interval '1 microsecond' WHERE id=$1`, job.ID)
	err = service.ProcessPasswordReset(ctx, auth.QueuedResetRequest{ID: job.ID, Lease: job.Lease, Email: content.Email})
	if !errors.Is(err, notification.ErrLeaseLost) {
		t.Fatalf("expired input issued a fresh reset: %v", err)
	}
	outboxCheck(t, ctx, s, `SELECT NOT EXISTS(SELECT 1 FROM password_resets)
		AND NOT EXISTS(SELECT 1 FROM mail_jobs WHERE kind='password_reset')
		AND EXISTS(SELECT 1 FROM mail_jobs WHERE id=$1 AND lease_token=$2 AND leased_until>clock_timestamp())`, job.ID, job.Lease)
}

type outboxSender struct {
	failed     bool
	next       notification.Sender
	deliveries []notification.Delivery
}

func (s *outboxSender) Deliver(ctx context.Context, delivery notification.Delivery) error {
	s.deliveries = append(s.deliveries, delivery)
	if s.failed {
		return errors.New("temporary fixture transport failure")
	}
	return s.next.Deliver(ctx, delivery)
}

func TestAuthOutboxWorkerRetrySurvivesReconstructionWithLocalMailer(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s.AuthRepository, auth.Options{})
	if err := service.Register(ctx, auth.RegisterInput{Email: "restart@example.test", Login: "restart", Password: "test-password"}); err != nil {
		t.Fatal(err)
	}
	original := outboxMessage(t, ctx, s, notification.Verification)
	content := outboxContent(t, original)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	failing := &outboxSender{failed: true}
	worker := notification.NewWorker(db.NewMailQueue(s.pool), testCipher(), failing, nil, logger)
	if worked, err := worker.ProcessOne(ctx); err != nil || !worked {
		t.Fatalf("first attempt: worked=%v err=%v", worked, err)
	}
	outboxCheck(t, ctx, s, `SELECT attempts=1 AND failed_at IS NULL AND lease_token IS NULL AND available_at>clock_timestamp()+interval '3 seconds' AND payload=$2 AND expires_at=$3 FROM mail_jobs WHERE id=$1`, original.ID, original.Ciphertext, original.ExpiresAt)
	if worked, err := worker.ProcessOne(ctx); err != nil || worked {
		t.Fatalf("retry ignored persisted backoff: worked=%v err=%v", worked, err)
	}
	directory := t.TempDir()
	local, err := mailer.New(mailer.Options{Driver: "local", Directory: directory, PublicURL: "https://game.example.test", From: "fixture@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	resumed := &outboxSender{next: local}
	worker = notification.NewWorker(db.NewMailQueue(s.pool), testCipher(), resumed, nil, logger)
	mustExec(t, ctx, s, `UPDATE mail_jobs SET available_at=clock_timestamp()-interval '1 second' WHERE id=$1`, original.ID)
	if worked, err := worker.ProcessOne(ctx); err != nil || !worked {
		t.Fatalf("reconstructed attempt: worked=%v err=%v", worked, err)
	}
	if len(failing.deliveries) != 1 || len(resumed.deliveries) != 1 {
		t.Fatal("unexpected delivery count")
	}
	first, second := failing.deliveries[0], resumed.deliveries[0]
	if first.ID != original.ID || second.ID != first.ID || first.Token != content.Token || second.Token != first.Token || !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatal("retry changed stable delivery identity, token or date")
	}
	outboxCheck(t, ctx, s, `SELECT NOT EXISTS(SELECT 1 FROM mail_jobs) AND EXISTS(SELECT 1 FROM email_verifications WHERE token=$1)`, auth.TokenDigest(content.Token))
	files, err := filepath.Glob(filepath.Join(directory, "*.eml"))
	if err != nil || len(files) != 1 {
		t.Fatal("reconstructed worker did not deliver one local letter")
	}
	file, err := os.Open(files[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	letter, err := mail.ReadMessage(file)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(quotedprintable.NewReader(letter.Body))
	if err != nil || !strings.Contains(string(body), content.Token) || letter.Header.Get("Message-ID") != "<"+original.ID+"@game.example.test>" || letter.Header.Get("Date") != first.CreatedAt.UTC().Format(time.RFC1123Z) {
		t.Fatal("local retry letter changed token, Message-ID or Date")
	}
}

func TestMailCleanupRetentionAndLeases(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `INSERT INTO mail_jobs(id,kind,key_id,payload,user_id,expires_at,failed_at,lease_token,leased_until)
		VALUES
		(repeat('1',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()-interval '1 second',NULL,NULL,NULL),
		(repeat('2',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()+interval '1 day',clock_timestamp()-interval '25 hours',NULL,NULL),
		(repeat('3',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()+interval '1 day',clock_timestamp()-interval '23 hours',NULL,NULL),
		(repeat('4',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()-interval '1 second',NULL,repeat('b',64),clock_timestamp()+interval '1 minute'),
		(repeat('5',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()+interval '1 day',NULL,NULL,NULL),
		(repeat('6',64),'password_changed',repeat('a',64),'ciphertext',1,clock_timestamp()-interval '1 second',NULL,NULL,NULL)`)
	preview, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1, DryRun: true})
	if err != nil || preview.Batches != 0 || preview.MailJobs != 3 {
		t.Fatalf("preview: %+v %v", preview, err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT id FROM mail_jobs WHERE id=repeat('6',64) FOR UPDATE`); err != nil {
		t.Fatal(err)
	}
	result, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	if err := s.pool.QueryRow(ctx, `SELECT array_agg(left(id,1) ORDER BY id) FROM mail_jobs`).Scan(&ids); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, "") != "3456" || result.MailJobs != 2 {
		t.Fatalf("expired/failed jobs not cleaned, or live/leased/locked jobs lost: %v (%+v)", ids, result)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE mail_jobs SET leased_until=clock_timestamp()-interval '1 second' WHERE id=repeat('4',64)`)
	if _, err := maintenance.Run(ctx, s.pool, maintenance.Options{BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT array_agg(left(id,1) ORDER BY id) FROM mail_jobs`).Scan(&ids); err != nil || strings.Join(ids, "") != "35" {
		t.Fatalf("unlocked/abandoned expired mail not cleaned: %v %v", ids, err)
	}
}
