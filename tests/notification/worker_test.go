package notification_test

import (
	"bytes"
	"context"
	"errors"
	"gotest/internal/notification"
	"log/slog"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"
)

type memoryStore struct {
	job               *notification.Job
	owns              bool
	ownsErr, errorAck error
	retries, acks     int
	delay             time.Duration
	permanent         bool
	mu                sync.Mutex
}

func (s *memoryStore) Claim(ctx context.Context) (*notification.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.job, nil
}
func (s *memoryStore) Owns(context.Context, notification.Job) (bool, error) { return s.owns, s.ownsErr }
func (s *memoryStore) Complete(context.Context, notification.Job) error {
	s.acks++
	if s.errorAck == nil {
		s.job = nil
	}
	return s.errorAck
}
func (s *memoryStore) Retry(_ context.Context, _ notification.Job, d time.Duration, p bool) error {
	s.retries++
	s.delay = d
	s.permanent = p
	return nil
}

type senderFunc func(context.Context, notification.Delivery) error

func (f senderFunc) Deliver(c context.Context, d notification.Delivery) error { return f(c, d) }
func queuedJob(t *testing.T, c *notification.Cipher, k notification.Kind) *notification.Job {
	t.Helper()
	token := ""
	if k == notification.Verification || k == notification.PasswordReset {
		token = strings.Repeat("a", 64)
	}
	m, err := c.Seal(k, "user@example.com", token, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return &notification.Job{Message: m, Lease: "lease", Attempts: 1, CreatedAt: time.Now().UTC()}
}

func TestWorkerRetriesTransientThenAcknowledgesSuccess(t *testing.T) {
	c := newCipher(t)
	s := &memoryStore{job: queuedJob(t, c, notification.Verification), owns: true}
	calls := 0
	id := s.job.ID
	created := s.job.CreatedAt
	send := senderFunc(func(_ context.Context, d notification.Delivery) error {
		calls++
		if d.ID != id || !d.CreatedAt.Equal(created) || d.Token != strings.Repeat("a", 64) {
			t.Fatal("delivery changed")
		}
		if calls == 1 {
			return &textproto.Error{Code: 451, Msg: "recipient secret"}
		}
		return nil
	})
	var logs bytes.Buffer
	w := notification.NewWorker(s, c, send, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
	if ok, err := w.ProcessOne(context.Background()); !ok || err != nil {
		t.Fatalf("first: %v %v", ok, err)
	}
	if s.retries != 1 || s.delay != 5*time.Second || s.permanent || s.acks != 0 {
		t.Fatal("bad retry")
	}
	s.job.Attempts++
	if _, err := w.ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.acks != 1 || calls != 2 {
		t.Fatal("not acknowledged")
	}
	if strings.Contains(logs.String(), "recipient secret") || strings.Contains(logs.String(), "user@example.com") {
		t.Fatal("unsafe logs")
	}
}
func TestWorkerPermanentFailuresAndAttemptLimit(t *testing.T) {
	for _, mode := range []string{"smtp", "decode", "limit"} {
		t.Run(mode, func(t *testing.T) {
			c := newCipher(t)
			s := &memoryStore{job: queuedJob(t, c, notification.Verification), owns: true}
			calls := 0
			if mode == "decode" {
				s.job.Ciphertext = []byte{1}
			}
			if mode == "limit" {
				s.job.Attempts = 8
			}
			w := notification.NewWorker(s, c, senderFunc(func(context.Context, notification.Delivery) error {
				calls++
				if mode == "smtp" {
					return &textproto.Error{Code: 550, Msg: "secret"}
				}
				return errors.New("transient")
			}), nil, nil)
			if _, err := w.ProcessOne(context.Background()); err != nil {
				t.Fatal(err)
			}
			if s.retries != 1 || !s.permanent || s.acks != 0 {
				t.Fatal("not permanent")
			}
			if mode == "decode" && calls != 0 {
				t.Fatal("corrupt mail sent")
			}
		})
	}
}
func TestWorkerDoesNotAcknowledgeStaleOrAtomicallyConsumedJobs(t *testing.T) {
	for _, mode := range []string{"stale", "lost", "reset", "ackfail", "expired"} {
		t.Run(mode, func(t *testing.T) {
			c := newCipher(t)
			kind := notification.Verification
			if mode == "reset" {
				kind = notification.ResetRequest
			}
			s := &memoryStore{job: queuedJob(t, c, kind), owns: mode != "stale" && mode != "expired"}
			calls, processed := 0, 0
			if mode == "expired" {
				m, _ := c.Seal(kind, "user@example.com", strings.Repeat("a", 64), time.Now().Add(-time.Hour))
				s.job.Message = m
			}
			if mode == "ackfail" {
				s.errorAck = errors.New("db failure")
			}
			w := notification.NewWorker(s, c, senderFunc(func(context.Context, notification.Delivery) error {
				calls++
				if mode == "lost" {
					return notification.ErrLeaseLost
				}
				return nil
			}), func(context.Context, notification.Job, notification.Content) error { processed++; return nil }, nil)
			_, err := w.ProcessOne(context.Background())
			if mode == "ackfail" {
				if err == nil || s.acks != 1 || s.retries != 0 {
					t.Fatal("ack failure retried")
				}
				return
			}
			if err != nil && !errors.Is(err, notification.ErrLeaseLost) {
				t.Fatal(err)
			}
			if mode == "reset" {
				if processed != 1 || s.acks != 0 || calls != 0 {
					t.Fatal("reset double ack")
				}
			} else if mode == "expired" {
				if calls != 0 {
					t.Fatal("expired mail sent")
				}
			} else if s.acks != 0 || s.retries != 0 {
				t.Fatal("stale work mutated")
			}
			if mode == "stale" && calls != 0 {
				t.Fatal("stale lease sent mail")
			}
		})
	}
}

func TestWorkerUsesStoreClockForExpiryDespiteApplicationClockSkew(t *testing.T) {
	c := newCipher(t)
	message, err := c.Seal(notification.Verification, "user@example.com", strings.Repeat("a", 64), time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	s := &memoryStore{job: &notification.Job{Message: message, Lease: "live-database-lease", Attempts: 1, CreatedAt: time.Now()}, owns: true}
	sent := false
	w := notification.NewWorker(s, c, senderFunc(func(context.Context, notification.Delivery) error { sent = true; return nil }), nil, nil)
	if _, err := w.ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !sent || s.acks != 1 {
		t.Fatal("application clock discarded a database-valid job")
	}
}

func TestWorkerBackoffCapsAndStopsAfterEighthFailure(t *testing.T) {
	c := newCipher(t)
	s := &memoryStore{job: queuedJob(t, c, notification.Verification), owns: true}
	w := notification.NewWorker(s, c, senderFunc(func(context.Context, notification.Delivery) error { return errors.New("temporary failure") }), nil, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	for i, want := range []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second, 40 * time.Second, 80 * time.Second, 160 * time.Second, 300 * time.Second, 300 * time.Second} {
		s.job.Attempts = i + 1
		if _, err := w.ProcessOne(context.Background()); err != nil {
			t.Fatal(err)
		}
		if s.delay != want || s.permanent != (i == 7) {
			t.Fatalf("attempt %d: delay=%v permanent=%v", i+1, s.delay, s.permanent)
		}
	}
}

func TestWorkerLostResetLeaseAndOwnershipErrorLeaveJobUntouched(t *testing.T) {
	for _, ownershipError := range []bool{false, true} {
		c := newCipher(t)
		s := &memoryStore{job: queuedJob(t, c, notification.ResetRequest), owns: true}
		processed := false
		if ownershipError {
			s.ownsErr = errors.New("database unavailable")
		}
		w := notification.NewWorker(s, c, nil, func(context.Context, notification.Job, notification.Content) error {
			processed = true
			return notification.ErrLeaseLost
		}, nil)
		if _, err := w.ProcessOne(context.Background()); err == nil {
			t.Fatal("expected failure")
		}
		if s.acks != 0 || s.retries != 0 || processed == ownershipError {
			t.Fatal("ownership failure mutated job")
		}
	}
}
func TestWorkerCancellationLeavesWorkAndJoinsLoops(t *testing.T) {
	c := newCipher(t)
	s := &memoryStore{job: queuedJob(t, c, notification.Verification), owns: true}
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 2)
	w := notification.NewWorker(s, c, senderFunc(func(ctx context.Context, _ notification.Delivery) error {
		started <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	}), nil, nil)
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker failed to stop")
	}
	if s.retries != 0 || s.acks != 0 {
		t.Fatal("cancel mutated work")
	}
}
