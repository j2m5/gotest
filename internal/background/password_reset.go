// Package background owns bounded, process-local jobs with explicit lifetimes.
package background

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"gotest/internal/auth"
	"gotest/internal/observability"
)

type PasswordResetService interface {
	RequestPasswordReset(context.Context, string) error
}

// PasswordResetQueue accepts requests before account lookup or mail delivery.
// It is intentionally not durable; shutdown cancels work and callers can retry.
type PasswordResetQueue struct {
	ctx      context.Context
	cancel   context.CancelFunc
	requests chan string
	workers  sync.WaitGroup
}

func NewPasswordResetQueue(parent context.Context, service PasswordResetService, logger *slog.Logger) *PasswordResetQueue {
	ctx, cancel := context.WithCancel(parent)
	q := &PasswordResetQueue{ctx: ctx, cancel: cancel, requests: make(chan string, 64)}
	if logger == nil {
		logger = slog.Default()
	}
	for range 2 {
		q.workers.Go(func() {
			for {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case email := <-q.requests:
					job, stop := context.WithTimeout(ctx, 70*time.Second)
					err := service.RequestPasswordReset(job, email)
					stop()
					if err != nil {
						logger.LogAttrs(ctx, slog.LevelError, "password_reset_request_failed", observability.ErrorAttrs(err)...)
					}
				}
			}
		})
	}
	return q
}

func (q *PasswordResetQueue) RequestPasswordReset(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	email, err := auth.ValidatePasswordResetEmail(email)
	if err != nil {
		return err
	}
	if q.ctx.Err() != nil {
		return auth.ErrResetUnavailable
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.ctx.Done():
		return auth.ErrResetUnavailable
	case q.requests <- email:
		return nil
	default:
		return auth.ErrResetUnavailable
	}
}

func (q *PasswordResetQueue) Close() {
	q.cancel()
	q.workers.Wait()
}
