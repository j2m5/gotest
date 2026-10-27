package background_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/background"
)

type blockedReset struct{ entered chan string }

func (s *blockedReset) RequestPasswordReset(ctx context.Context, email string) error {
	s.entered <- email
	<-ctx.Done()
	return ctx.Err()
}

func TestResetQueueAcceptsBeforeLookupAndBoundsWork(t *testing.T) {
	service := &blockedReset{entered: make(chan string, 2)}
	queue := background.NewPasswordResetQueue(context.Background(), service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer queue.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, email := range []string{" Known@Example.test ", "unknown@example.test"} {
		if err := queue.RequestPasswordReset(ctx, email); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		select {
		case email := <-service.entered:
			if email != "known@example.test" && email != "unknown@example.test" {
				t.Fatal("email not normalized")
			}
		case <-ctx.Done():
			t.Fatal("worker not started")
		}
	}
	for range 64 {
		if err := queue.RequestPasswordReset(ctx, "queued@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	if err := queue.RequestPasswordReset(ctx, "overflow@example.test"); !errors.Is(err, auth.ErrResetUnavailable) {
		t.Fatalf("unbounded queue: %v", err)
	}
	queue.Close() // Must cancel in-flight requests and join workers without executing the backlog.
	if err := queue.RequestPasswordReset(ctx, "closed@example.test"); !errors.Is(err, auth.ErrResetUnavailable) {
		t.Fatal("closed queue accepted work")
	}
}

func TestResetQueueRejectsInvalidEmailAndCancelledRequest(t *testing.T) {
	service := &blockedReset{entered: make(chan string, 2)}
	queue := background.NewPasswordResetQueue(context.Background(), service, nil)
	defer queue.Close()
	if err := queue.RequestPasswordReset(context.Background(), "invalid email"); err == nil {
		t.Fatal("invalid email queued")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := queue.RequestPasswordReset(ctx, "cancelled@example.test"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled request accepted")
	}
	select {
	case <-service.entered:
		t.Fatal("rejected request reached account lookup")
	default:
	}
}
