package combat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gotest/internal/combat"
)

type blockingCombatStore struct {
	entered chan struct{}
	active  atomic.Int32
	max     atomic.Int32
}

type deferredCombatStore struct {
	calls   atomic.Int32
	healthy chan struct{}
	retryAt time.Time
}

func (s *deferredCombatStore) ProcessDue(ctx context.Context) (bool, error) {
	if s.calls.Add(1) <= 2 {
		return true, &combat.ProcessingError{BattleID: 42, CharacterID: 7, RetryAt: s.retryAt,
			Err: errors.New("sensitive test value: password=private")}
	}
	select {
	case s.healthy <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return false, ctx.Err()
}

func TestCombatWorkerContinuesAfterDeferredFailureAndLogsSafeContext(t *testing.T) {
	store := &deferredCombatStore{healthy: make(chan struct{}, 1), retryAt: time.Now().UTC().Add(time.Second)}
	var output bytes.Buffer
	worker := combat.NewWorker(store, slog.New(slog.NewJSONHandler(&output, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { worker.Run(ctx); close(done) }()
	healthy := false
	select {
	case <-store.healthy:
		healthy = true
	case <-time.After(750 * time.Millisecond):
	}
	cancel()
	<-done
	if !healthy {
		t.Fatal("isolated encounter errors paused healthy queue work")
	}
	if strings.Contains(output.String(), "password") || strings.Contains(output.String(), "sensitive") {
		t.Fatal("raw error leaked into worker log")
	}
	decoder := json.NewDecoder(&output)
	count := 0
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record["battle_id"] != float64(42) || record["character_id"] != float64(7) || record["retry_at"] != store.retryAt.Format(time.RFC3339Nano) || record["error_kind"] != "internal" {
			t.Fatal("missing safe failure context", record)
		}
		count++
	}
	if count != 2 {
		t.Fatal("deferred errors were not logged", count)
	}
}

func (s *blockingCombatStore) ProcessDue(ctx context.Context) (bool, error) {
	active := s.active.Add(1)
	defer s.active.Add(-1)
	for old := s.max.Load(); active > old; old = s.max.Load() {
		if s.max.CompareAndSwap(old, active) {
			break
		}
	}
	s.entered <- struct{}{}
	<-ctx.Done()
	return false, ctx.Err()
}
func TestCombatWorkerBoundsConcurrencyAndJoinsShutdown(t *testing.T) {
	store := &blockingCombatStore{entered: make(chan struct{}, 8)}
	worker := combat.NewWorker(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { worker.Run(ctx); close(done) }()
	for range 2 {
		select {
		case <-store.entered:
		case <-time.After(2 * time.Second):
			t.Fatal("worker did not start")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("workers not joined")
	}
	if store.active.Load() != 0 || store.max.Load() != 2 {
		t.Fatal(store.active.Load(), store.max.Load())
	}
}

type failingCombatStore struct{ calls atomic.Int32 }

func (s *failingCombatStore) ProcessDue(context.Context) (bool, error) {
	s.calls.Add(1)
	return false, errors.New("database unavailable")
}
func TestCombatWorkerBacksOffAndCancellationInterruptsWait(t *testing.T) {
	store := &failingCombatStore{}
	worker := combat.NewWorker(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	worker.Run(ctx)
	if count := store.calls.Load(); count < 1 || count > 2 {
		t.Fatal("retry spin", count)
	}
}
