package combat

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"gotest/internal/observability"
)

type DueStore interface {
	// True includes a committed retry. An error with true reports an isolated
	// encounter failure; an error with false may require a global retry delay.
	ProcessDue(context.Context) (bool, error)
}

// ProcessingError describes an encounter whose effects were rolled back and
// whose next retry was committed. Err is for classification, never raw logging.
type ProcessingError struct {
	BattleID    int64
	CharacterID int
	RetryAt     time.Time
	Err         error
}

func (e *ProcessingError) Error() string { return "combat: processing deferred" }
func (e *ProcessingError) Unwrap() error { return e.Err }

type Worker struct {
	store  DueStore
	logger *slog.Logger
}

func NewWorker(store DueStore, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{store: store, logger: logger}
}

// Run owns a fixed pool and waits for all workers before returning. Database
// row locks serialize multiple application processes as well as these goroutines.
func (w *Worker) Run(ctx context.Context) {
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() {
			for ctx.Err() == nil {
				operation, cancel := context.WithTimeout(ctx, 5*time.Second)
				worked, err := w.store.ProcessDue(operation)
				cancel()
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					attrs := observability.ErrorAttrs(err)
					var deferred *ProcessingError
					if errors.As(err, &deferred) {
						attrs = append(attrs, slog.Int64("battle_id", deferred.BattleID),
							slog.Int("character_id", deferred.CharacterID), slog.Time("retry_at", deferred.RetryAt))
					}
					w.logger.LogAttrs(ctx, slog.LevelWarn, "combat processing failed", attrs...)
				}
				if worked {
					// A deferred encounter already has its own persisted delay.
					// Keep draining healthy work without a global error pause.
					continue
				}
				delay := 250 * time.Millisecond
				if err != nil {
					delay = time.Second
				}
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		})
	}
	workers.Wait()
}
