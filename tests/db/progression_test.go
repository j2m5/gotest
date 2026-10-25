package db_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"gotest/internal/db"
	"gotest/internal/progression"
)

func TestProgressionRepositoryGrantsCumulativeExperience(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1; UPDATE characters SET hp_current=7,mp_current=3 WHERE id=1`)
	unchanged := progressionUnrelatedSnapshot(t, ctx, s)
	var original *progression.Result
	for _, tt := range []struct {
		key           string
		amount        int64
		before, after progression.State
	}{
		{"quest:1", 99, progression.State{Level: 1, Experience: 0}, progression.State{Level: 1, Experience: 99}},
		{"quest:2", 1, progression.State{Level: 1, Experience: 99}, progression.State{Level: 2, Experience: 100}},
		{"quest:3", 150, progression.State{Level: 2, Experience: 100}, progression.State{Level: 3, Experience: 250}},
	} {
		reward := progression.Reward{CharacterID: 1, Key: tt.key, Experience: tt.amount}
		result, err := service.Grant(ctx, reward)
		want := progression.Result{CharacterID: 1, Key: tt.key, Requested: tt.amount, Change: progression.Change{Before: tt.before, After: tt.after, Granted: tt.amount}}
		if err != nil || result == nil || *result != want {
			t.Fatalf("Grant = %+v, %v; want %+v", result, err, want)
		}
		if original == nil {
			original = result
		}
		assertProgressionState(t, ctx, s, 1, tt.after)
		if got := progressionUnrelatedSnapshot(t, ctx, s); got != unchanged {
			t.Fatal("grant changed stats, selection or another character")
		}
	}
	// Replay returns the original outcome even after other rewards changed the character.
	reward := progression.Reward{CharacterID: 1, Key: "quest:1", Experience: 99}
	replay, err := service.Grant(ctx, reward)
	wantReplay := *original
	wantReplay.Replayed = true
	if err != nil || replay == nil || *replay != wantReplay {
		t.Fatalf("historical replay = %+v, %v; want %+v", replay, err, wantReplay)
	}
	beforeConflict := characterStatsSnapshot(t, ctx, s)
	reward.Experience++
	if _, err := service.Grant(ctx, reward); !errors.Is(err, progression.ErrRewardConflict) {
		t.Fatalf("amount conflict: %v", err)
	}
	if characterStatsSnapshot(t, ctx, s) != beforeConflict {
		t.Fatal("conflicting reward changed state")
	}
	assertRewardCount(t, ctx, s, 3)
	// The same source event can reward a different participant independently.
	reward.CharacterID, reward.Experience = 2, 99
	result, err := service.Grant(ctx, reward)
	if err != nil || result == nil || result.Replayed {
		t.Fatalf("other participant = %+v, %v", result, err)
	}
	assertProgressionState(t, ctx, s, 2, progression.State{Level: 1, Experience: 99})
	assertRewardCount(t, ctx, s, 4)
}

func TestProgressionRepositoryCapsAndRecordsZeroAwards(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	result, err := service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:huge", Experience: math.MaxInt64})
	if err != nil || result == nil || result.Change.Granted != 154200 || result.Change.After != (progression.State{Level: 50, Experience: 154200}) {
		t.Fatalf("capped award = %+v, %v", result, err)
	}
	before := characterStatsSnapshot(t, ctx, s)
	zero := progression.Reward{CharacterID: 1, Key: "quest:at-cap", Experience: 1000}
	result, err = service.Grant(ctx, zero)
	if err != nil || result == nil || result.Change.Granted != 0 || result.Replayed {
		t.Fatalf("max award = %+v, %v", result, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("max award changed character or its updated_at")
	}
	assertRewardCount(t, ctx, s, 2)
	// A higher threshold for the existing maximum does not restart earning XP.
	mustExec(t, ctx, s, `UPDATE levels SET experience_required=180000 WHERE level=50`)
	result, err = service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:raised-cap", Experience: 1000})
	if err != nil || result == nil || result.Change.Granted != 0 {
		t.Fatalf("changed max threshold = %+v, %v", result, err)
	}
	// Adding another level only enables new rewards; capped events cannot be banked.
	mustExec(t, ctx, s, `INSERT INTO levels(level,experience_required) VALUES(51,200000)`)
	result, err = service.Grant(ctx, zero)
	if err != nil || result == nil || !result.Replayed || result.Change.Granted != 0 {
		t.Fatalf("replay after extension = %+v, %v", result, err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 50, Experience: 154200})
	result, err = service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:new-level", Experience: 1000})
	if err != nil || result == nil || result.Change.After != (progression.State{Level: 50, Experience: 155200}) {
		t.Fatalf("new reward after extension = %+v, %v", result, err)
	}
	assertRewardCount(t, ctx, s, 4)
}

func TestProgressionRepositoryConcurrentRewards(t *testing.T) {
	for _, sameKey := range []bool{true, false} {
		t.Run(fmt.Sprintf("same_key_%t", sameKey), func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			service := progression.NewService(db.NewProgressionRepository(s.pool))
			type outcome struct {
				result *progression.Result
				err    error
			}
			const workers = 10
			start, done := make(chan struct{}), make(chan outcome, workers)
			for i := range workers {
				go func() {
					<-start
					key := "battle:shared"
					if !sameKey {
						key = fmt.Sprintf("battle:%d", i)
					}
					result, err := service.Grant(ctx, progression.Reward{CharacterID: 1, Key: key, Experience: 30})
					done <- outcome{result, err}
				}()
			}
			close(start)
			fresh := 0
			for range workers {
				got := <-done
				if got.err != nil || got.result == nil {
					t.Fatalf("concurrent grant = %+v, %v", got.result, got.err)
				}
				if !got.result.Replayed {
					fresh++
				}
			}
			wantCount, wantState := workers, progression.State{Level: 3, Experience: 300}
			if sameKey {
				wantCount, wantState = 1, progression.State{Level: 1, Experience: 30}
			}
			if fresh != wantCount {
				t.Fatalf("fresh awards = %d; want %d", fresh, wantCount)
			}
			assertRewardCount(t, ctx, s, wantCount)
			assertProgressionState(t, ctx, s, 1, wantState)
		})
	}
}

func TestProgressionRepositoryReadsCurrentCurveAndReplaysWithoutIt(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	reward := progression.Reward{CharacterID: 1, Key: "quest:old-balance", Experience: 100}
	original, err := service.Grant(ctx, reward)
	if err != nil {
		t.Fatal(err)
	}
	// Intermediate thresholds change without restarting the service.
	mustExec(t, ctx, s, `UPDATE levels SET experience_required=200 WHERE level=2`)
	_, err = service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:new-balance", Experience: 50})
	if err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 2, Experience: 150})
	// Malformed admin data blocks fresh grants atomically, but not historical replay.
	mustExec(t, ctx, s, `UPDATE levels SET experience_required=999 WHERE level=2`)
	before := characterStatsSnapshot(t, ctx, s)
	_, err = service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:bad-curve", Experience: 50})
	if !errors.Is(err, progression.ErrInvalidCurve) {
		t.Fatalf("invalid curve = %v", err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("invalid curve partially applied a grant")
	}
	assertRewardCount(t, ctx, s, 2)
	result, err := service.Grant(ctx, reward)
	want := *original
	want.Replayed = true
	if err != nil || result == nil || *result != want {
		t.Fatalf("replay with invalid catalog = %+v, %v", result, err)
	}
	// Lowered thresholds affect the next grant, without lowering existing XP.
	mustExec(t, ctx, s, `UPDATE levels SET experience_required=10 WHERE level=2; UPDATE levels SET experience_required=20 WHERE level=3`)
	_, err = service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:lowered-balance", Experience: 1})
	if err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 3, Experience: 151})
}

func TestProgressionRepositoryRollsBackWhenLedgerWriteFails(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	mustExec(t, ctx, s, `CREATE FUNCTION reject_test_reward() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'test reward failure'; END $$;
		CREATE TRIGGER reject_test_reward BEFORE INSERT ON character_experience_rewards
		FOR EACH ROW EXECUTE FUNCTION reject_test_reward()`)
	before := characterStatsSnapshot(t, ctx, s)
	reward := progression.Reward{CharacterID: 1, Key: "quest:retry", Experience: 250}
	if _, err := service.Grant(ctx, reward); err == nil {
		t.Fatal("failing ledger accepted grant")
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("failed ledger left changed character/timestamp")
	}
	assertRewardCount(t, ctx, s, 0)
	mustExec(t, ctx, s, `DROP TRIGGER reject_test_reward ON character_experience_rewards`)
	result, err := service.Grant(ctx, reward)
	if err != nil || result == nil || result.Replayed {
		t.Fatalf("retry = %+v, %v", result, err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 3, Experience: 250})
	assertRewardCount(t, ctx, s, 1)
}

func TestProgressionRepositoryRejectsInvalidMissingAndCancelledGrants(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewProgressionRepository(s.pool)
	before := characterStatsSnapshot(t, ctx, s)
	for _, reward := range []progression.Reward{{CharacterID: 1, Key: "bad key", Experience: 10}, {CharacterID: 1, Key: "ok", Experience: 0}} {
		if _, err := r.Grant(ctx, reward); !errors.Is(err, progression.ErrInvalidReward) {
			t.Fatalf("invalid direct repository input = %v", err)
		}
	}
	if _, err := r.Grant(ctx, progression.Reward{CharacterID: 999999, Key: "quest:missing", Experience: 100}); !errors.Is(err, progression.ErrNotFound) {
		t.Fatalf("missing character = %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := r.Grant(cancelled, progression.Reward{CharacterID: 1, Key: "quest:cancelled", Experience: 100}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled input = %v", err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("rejected input changed state")
	}
	assertRewardCount(t, ctx, s, 0)
}

func TestProgressionRepositoryCancellationWhileWaitingDoesNotBlockOtherCharacters(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewProgressionRepository(s.pool)
	blocker, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback(ctx)
	var pid int
	if err := blocker.QueryRow(ctx, `SELECT pg_backend_pid() FROM characters WHERE id=1 FOR UPDATE`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := r.Grant(waitCtx, progression.Reward{CharacterID: 1, Key: "quest:wait", Experience: 100})
		done <- err
	}()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("grant did not wait for character: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
	// The other character of this account remains independent of the locked one.
	otherCtx, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()
	if _, err := r.Grant(otherCtx, progression.Reward{CharacterID: 2, Key: "quest:independent", Experience: 100}); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled waiter = %v", err)
	}
	if err := blocker.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 1, Experience: 0})
	assertProgressionState(t, ctx, s, 2, progression.State{Level: 2, Experience: 100})
	assertRewardCount(t, ctx, s, 1)
	// Cancellation releases its transaction; a later attempt with the same key works.
	if _, err := r.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:wait", Experience: 100}); err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 2, Experience: 100})
}

func assertProgressionState(t *testing.T, ctx context.Context, s *testStorage, id int, want progression.State) {
	t.Helper()
	var got progression.State
	if err := s.pool.QueryRow(ctx, `SELECT level,experience FROM characters WHERE id=$1`, id).Scan(&got.Level, &got.Experience); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("character %d = %+v; want %+v", id, got, want)
	}
}

func assertRewardCount(t *testing.T, ctx context.Context, s *testStorage, want int) {
	t.Helper()
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM character_experience_rewards`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("ledger entries = %d; want %d", count, want)
	}
}

func progressionUnrelatedSnapshot(t *testing.T, ctx context.Context, s *testStorage) string {
	t.Helper()
	var snapshot string
	if err := s.pool.QueryRow(ctx, `SELECT jsonb_build_object(
		'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM users u),
		'characters',(SELECT jsonb_agg(CASE WHEN id=1 THEN to_jsonb(c)-ARRAY['level','experience','updated_at'] ELSE to_jsonb(c) END ORDER BY id) FROM characters c),
		'abilities',(SELECT jsonb_agg(to_jsonb(a) ORDER BY character_id,ability_id) FROM character_abilities a))::text`).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}
