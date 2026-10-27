package notification

import (
	"context"
	"errors"
	"gotest/internal/observability"
	"log/slog"
	"net/textproto"
	"sync"
	"time"
)

type Worker struct {
	store  Store
	cipher *Cipher
	sender Sender
	reset  ResetProcessor
	logger *slog.Logger
}

func NewWorker(store Store, cipher *Cipher, sender Sender, reset ResetProcessor, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{store: store, cipher: cipher, sender: sender, reset: reset, logger: logger}
}

// ProcessOne handles one lease. Acknowledgement errors deliberately leave the
// lease untouched: delivery can be repeated after reclaim, but cannot be lost.
func (w *Worker) ProcessOne(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, 70*time.Second)
	defer cancel()
	job, err := w.store.Claim(ctx)
	if err != nil {
		return false, err
	}
	if job == nil {
		return false, nil
	}
	owned, err := w.store.Owns(ctx, *job)
	if err != nil {
		return true, err
	}
	if !owned {
		return true, ErrLeaseLost
	}
	// Owns checks expiry using the persistence clock. An application clock
	// ahead of PostgreSQL must not discard work that is still valid there.
	content, err := w.cipher.Open(job.Message)
	if err != nil {
		return true, w.fail(ctx, *job, err, true)
	}
	// The decode can be expensive relative to cancellation; do not start work
	// once shutdown or the bounded job deadline has been reached.
	if err := ctx.Err(); err != nil {
		return true, err
	}
	if job.Kind == ResetRequest {
		if w.reset == nil {
			err = errors.New("notification reset processor unavailable")
		} else {
			err = w.reset(ctx, *job, content)
		}
	} else if w.sender == nil {
		err = errors.New("notification sender unavailable")
	} else {
		err = w.sender.Deliver(ctx, Delivery{ID: job.ID, Kind: job.Kind, Email: content.Email, Token: content.Token, CreatedAt: job.CreatedAt})
	}
	if err != nil {
		if errors.Is(err, ErrLeaseLost) || errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return true, err
		}
		var smtp *textproto.Error
		permanent := errors.As(err, &smtp) && smtp.Code >= 500 && smtp.Code < 600
		return true, w.fail(ctx, *job, err, permanent)
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}
	if job.Kind == ResetRequest {
		return true, nil
	}
	return true, w.store.Complete(ctx, *job)
}

func (w *Worker) fail(ctx context.Context, job Job, cause error, permanent bool) error {
	if errors.Is(cause, ErrLeaseLost) || ctx.Err() != nil {
		return cause
	}
	delay := 5 * time.Second
	for attempt := 1; attempt < job.Attempts && delay < 5*time.Minute; attempt++ {
		delay *= 2
	}
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	attrs := []slog.Attr{slog.String("job_id", job.ID), slog.String("kind", string(job.Kind)), slog.Int("attempts", job.Attempts)}
	attrs = append(attrs, observability.ErrorAttrs(cause)...)
	w.logger.LogAttrs(ctx, slog.LevelWarn, "notification attempt failed", attrs...)
	return w.store.Retry(ctx, job, delay, permanent || job.Attempts >= MaxAttempts)
}

// Run blocks until cancellation and joins both workers before returning.
func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			for ctx.Err() == nil {
				worked, err := w.ProcessOne(ctx)
				if ctx.Err() != nil {
					return
				}
				if err != nil && !errors.Is(err, ErrLeaseLost) {
					w.logger.LogAttrs(ctx, slog.LevelWarn, "notification processing failed", observability.ErrorAttrs(err)...)
				}
				if !worked || err != nil {
					timer := time.NewTimer(250 * time.Millisecond)
					select {
					case <-ctx.Done():
						timer.Stop()
						return
					case <-timer.C:
					}
				}
			}
		})
	}
	wg.Wait()
}
