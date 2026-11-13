package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/combat"
	"gotest/internal/db"
)

type loadConfig struct {
	Battles  int
	Mixed    bool
	Duration time.Duration
}

type loadResult struct {
	Profile             string                    `json:"profile"`
	Battles             int                       `json:"battles"`
	Go                  string                    `json:"go"`
	OS                  string                    `json:"os"`
	CPUs                int                       `json:"cpus"`
	PostgreSQL          string                    `json:"postgresql"`
	PoolLimit           int32                     `json:"pool_limit"`
	SetupSeconds        float64                   `json:"setup_seconds"`
	WindowSeconds       float64                   `json:"window_seconds"`
	HTTPDrainSeconds    float64                   `json:"http_drain_seconds"`
	NPCEvents           int                       `json:"npc_events"`
	NPCAttackedBattles  int                       `json:"npc_attacked_battles"`
	NPCLagMS            loadQuantiles             `json:"npc_insert_lag_ms"`
	WorkerTransactions  int64                     `json:"worker_transactions"`
	WorkerErrors        int64                     `json:"worker_errors"`
	MaxDueBattles       int                       `json:"max_due_battles"`
	DueBattlesAtStop    int                       `json:"due_battles_at_stop"`
	MaxDueAgeMS         float64                   `json:"max_due_age_ms"`
	PoolEmptyAcquires   int64                     `json:"pool_empty_acquires"`
	PoolCanceledAcquire int64                     `json:"pool_canceled_acquires"`
	PoolAcquireMS       float64                   `json:"pool_acquire_total_ms"`
	Commands            int                       `json:"committed_commands"`
	CommandEvidence     loadCommandEvidence       `json:"command_evidence"`
	HTTP                map[string]*loadHTTPStats `json:"http"`
	commandOutcomes     map[string]loadCommandState
}

type loadDueProbe struct {
	repository *db.CombatRepository
	parent     context.Context
	worked     atomic.Int64
	failures   atomic.Int64
}

func (p *loadDueProbe) ProcessDue(ctx context.Context) (bool, error) {
	worked, err := p.repository.ProcessDue(ctx)
	if worked && err == nil {
		p.worked.Add(1)
	}
	if err != nil && p.parent.Err() == nil {
		p.failures.Add(1)
	}
	return worked, err
}

func runCombatLoad(t testing.TB, config loadConfig) loadResult {
	t.Helper()
	setupStarted := time.Now()
	s, _ := characterTestStorage(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+45*time.Second)
	defer cancel()
	actors := combatLoadFixture(t, s, ctx, config.Battles)
	result := loadResult{Profile: "burst", Battles: config.Battles, Go: runtime.Version(), OS: runtime.GOOS,
		CPUs: runtime.NumCPU(), PoolLimit: s.pool.Config().MaxConns}
	if config.Mixed {
		result.Profile = "mixed"
	}
	// A single observer connection keeps pool saturation from hiding its own queue.
	observerConfig := s.pool.Config()
	observerConfig.MaxConns, observerConfig.MinConns = 1, 0
	observer, err := pgxpool.NewWithConfig(ctx, observerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	if err := observer.QueryRow(ctx, `SHOW server_version`).Scan(&result.PostgreSQL); err != nil {
		t.Fatal(err)
	}
	repository := db.NewCombatRepository(s.pool)
	server := httptest.NewServer(combatLoadRouter(t, s, repository))
	defer server.Close()
	var startDB time.Time
	if err := observer.QueryRow(ctx, `SELECT clock_timestamp()+interval '1 second'`).Scan(&startDB); err != nil {
		t.Fatal(err)
	}
	start := time.Now().Add(time.Second)
	spread := 0
	if config.Mixed {
		spread = 2000
	}
	mustExec(t, ctx, s, `UPDATE battle_executions SET next_action_at=$1::timestamptz+
 ((battle_id-(SELECT min(battle_id) FROM battle_executions))*$2::double precision/$3)*interval '1 millisecond'`, startDB, spread, config.Battles)
	result.SetupSeconds = time.Since(setupStarted).Seconds()
	poolBefore := s.pool.Stat()
	workerCtx, stopWorker := context.WithCancel(ctx)
	probe := &loadDueProbe{repository: repository, parent: workerCtx}
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		combat.NewWorker(probe, slog.New(slog.NewTextHandler(io.Discard, nil))).Run(workerCtx)
	}()
	defer func() { stopWorker(); <-workerDone }()
	var metrics *loadHTTPMetrics
	joinHTTP := func() {}
	if config.Mixed {
		httpCtx, stopHTTP := context.WithCancel(ctx)
		metrics, joinHTTP = startCombatLoadHTTP(httpCtx, server.URL, actors, start, config.Duration)
		defer func() { stopHTTP(); joinHTTP() }()
	}
	if !waitCombatLoadUntil(ctx, start) {
		t.Fatal("load start deadline exceeded")
	}
	stopAt := start.Add(config.Duration)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		var due, attacked int
		var age float64
		err := observer.QueryRow(ctx, `SELECT
 count(*) FILTER (WHERE GREATEST(LEAST(next_action_at,expires_at),retry_at)<=statement_timestamp()),
 COALESCE(max(EXTRACT(EPOCH FROM statement_timestamp()-LEAST(next_action_at,expires_at))*1000)
 FILTER (WHERE GREATEST(LEAST(next_action_at,expires_at),retry_at)<=statement_timestamp()),0)::double precision,
 count(*) FILTER (WHERE pattern_step>0) FROM battle_executions WHERE next_action_at IS NOT NULL`).Scan(&due, &age, &attacked)
		if err != nil {
			t.Fatal("load queue observation", err)
		}
		result.MaxDueBattles = max(result.MaxDueBattles, due)
		result.MaxDueAgeMS = max(result.MaxDueAgeMS, age)
		result.DueBattlesAtStop = due
		if (!config.Mixed && attacked == config.Battles) || !time.Now().Before(stopAt) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("load observation timed out")
		case <-tick.C:
		}
	}
	var endDB time.Time
	if err := observer.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&endDB); err != nil {
		t.Fatal(err)
	}
	result.WindowSeconds = time.Since(start).Seconds()
	stopWorker()
	<-workerDone
	drainStarted := time.Now()
	finishCombatLoadHTTP(server, joinHTTP)
	result.HTTPDrainSeconds = time.Since(drainStarted).Seconds()
	if metrics != nil {
		result.HTTP = metrics.result()
		result.commandOutcomes = metrics.commandResults()
	}
	result.WorkerTransactions, result.WorkerErrors = probe.worked.Load(), probe.failures.Load()
	poolAfter := s.pool.Stat()
	result.PoolEmptyAcquires = poolAfter.EmptyAcquireCount() - poolBefore.EmptyAcquireCount()
	result.PoolCanceledAcquire = poolAfter.CanceledAcquireCount() - poolBefore.CanceledAcquireCount()
	result.PoolAcquireMS = float64(poolAfter.AcquireDuration()-poolBefore.AcquireDuration()) / float64(time.Millisecond)
	err = observer.QueryRow(ctx, `SELECT count(*),count(DISTINCT battle_id),
 COALESCE(percentile_disc(0.50) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM created_at-occurred_at)*1000),0)::double precision,
 COALESCE(percentile_disc(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM created_at-occurred_at)*1000),0)::double precision,
 COALESCE(percentile_disc(0.99) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM created_at-occurred_at)*1000),0)::double precision,
 COALESCE(max(EXTRACT(EPOCH FROM created_at-occurred_at)*1000),0)::double precision
 FROM battle_events WHERE payload->>'kind'='npc_ability' AND created_at>=$1 AND created_at<=$2`, startDB, endDB).Scan(
		&result.NPCEvents, &result.NPCAttackedBattles, &result.NPCLagMS.P50, &result.NPCLagMS.P95, &result.NPCLagMS.P99, &result.NPCLagMS.Max)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkCombatLoadIntegrity(ctx, observer, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func finishCombatLoadHTTP(server *httptest.Server, joinHTTP func()) {
	joinHTTP()
	// A client timeout can precede commit/rollback and handler return. Inspect
	// final database state only once the server has joined those handlers too.
	server.Close()
}

func checkCombatLoadIntegrity(ctx context.Context, pool *pgxpool.Pool, result *loadResult) error {
	var invalid, active, casts, rewards int
	err := pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM battle_executions e WHERE
 e.event_sequence<>(SELECT count(*) FROM battle_events v WHERE v.battle_id=e.battle_id)
 OR e.event_sequence<>(SELECT max(sequence) FROM battle_events v WHERE v.battle_id=e.battle_id)
 OR e.pattern_step<>(SELECT count(*) FROM battle_events v WHERE v.battle_id=e.battle_id AND payload->>'kind'='npc_ability')),
 (SELECT count(*) FROM battles WHERE status='active'),
 (SELECT count(*) FROM battle_commands),
 (SELECT count(*) FROM battle_events WHERE payload->>'kind'='player_ability'),
 (SELECT count(*) FROM wallet_entries WHERE source='pve')`).Scan(&invalid, &active, &result.Commands, &casts, &rewards)
	if err != nil {
		return err
	}
	if invalid != 0 || active != result.Battles || result.Commands != casts || rewards != 0 {
		return fmt.Errorf("load integrity: invalid=%d active=%d commands=%d casts=%d rewards=%d", invalid, active, result.Commands, casts, rewards)
	}
	if use := result.HTTP["use"]; use != nil {
		if err := checkLoadCommandCounts(result.Commands, use); err != nil {
			return err
		}
		if len(result.commandOutcomes) != use.Scheduled {
			return fmt.Errorf("load command outcomes missing: got %d, want %d", len(result.commandOutcomes), use.Scheduled)
		}
		rows, err := pool.Query(ctx, `SELECT command_key FROM battle_commands`)
		if err != nil {
			return err
		}
		committed := make(map[string]bool)
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err != nil {
				rows.Close()
				return err
			}
			if committed[key] {
				rows.Close()
				return fmt.Errorf("duplicate load command key %s", key)
			}
			committed[key] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		result.CommandEvidence, err = checkLoadCommandEvidence(result.commandOutcomes, committed)
		if err != nil {
			return err
		}
	}
	return checkCombatLoadResources(ctx, pool)
}

func checkLoadCommandCounts(committed int, use *loadHTTPStats) error {
	ambiguous := use.TransportErrors
	for code, count := range use.Statuses {
		if code >= 500 {
			ambiguous += count
		}
	}
	if committed < use.Statuses[200] || committed > use.Statuses[200]+ambiguous {
		return fmt.Errorf("load command accounting: committed=%d, HTTP=%+v", committed, use)
	}
	return nil
}

func TestCombatLoadHarness(t *testing.T) {
	result := runCombatLoad(t, loadConfig{Battles: 4, Mixed: true, Duration: 3 * time.Second})
	if result.NPCAttackedBattles != 4 || result.WorkerErrors != 0 || result.Commands == 0 || len(result.HTTP) != 4 {
		t.Fatalf("incomplete load scenario: %+v", result)
	}
	for kind, stats := range result.HTTP {
		if stats.Scheduled != stats.Completed || stats.Dropped != 0 || stats.TransportErrors != 0 || stats.Statuses[200] != stats.Completed {
			t.Fatalf("load harness %s: %+v", kind, stats)
		}
	}
}

func BenchmarkCombatLoad(b *testing.B) {
	duration := 15 * time.Second
	if value := os.Getenv("COMBAT_LOAD_DURATION"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed < 5*time.Second || parsed > time.Minute {
			b.Fatal("COMBAT_LOAD_DURATION must be between 5s and 1m")
		}
		duration = parsed
	}
	for _, mixed := range []bool{false, true} {
		profile := "burst"
		if mixed {
			profile = "mixed"
		}
		for _, count := range []int{100, 500, 1000} {
			b.Run(fmt.Sprintf("%s/%d", profile, count), func(b *testing.B) {
				if b.N != 1 {
					b.Fatal("combat load benchmarks require -benchtime=1x")
				}
				window := 30 * time.Second
				if mixed {
					window = duration
				}
				result := runCombatLoad(b, loadConfig{Battles: count, Mixed: mixed, Duration: window})
				encoded, err := json.Marshal(result)
				if err != nil {
					b.Fatal(err)
				}
				b.Logf("COMBAT_LOAD_RESULT %s", encoded)
				b.ReportMetric(0, "ns/op") // This is a whole scenario, not a single operation.
				b.ReportMetric(result.NPCLagMS.P95, "npc-p95-ms")
				b.ReportMetric(float64(result.NPCEvents)/result.WindowSeconds, "npc-events/s")
			})
		}
	}
}
