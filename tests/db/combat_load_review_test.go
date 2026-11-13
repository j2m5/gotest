package db_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCombatLoadCountsOperationTimeoutButNotShutdown(t *testing.T) {
	f := newExecutionFixture(t)
	probe := &loadDueProbe{repository: f.r, parent: f.ctx}
	operation, cancel := context.WithDeadline(f.ctx, time.Now().Add(-time.Second))
	defer cancel()
	if _, err := probe.ProcessDue(operation); err == nil || probe.failures.Load() != 1 {
		t.Fatal("operation timeout was hidden from load metrics", err, probe.failures.Load())
	}
	shutdown, stop := context.WithCancel(f.ctx)
	stop()
	probe.parent = shutdown
	_, _ = probe.ProcessDue(shutdown)
	if probe.failures.Load() != 1 {
		t.Fatal("intentional shutdown counted as a load failure")
	}
}

func TestCombatLoadJoinsHandlersAfterClientTimeout(t *testing.T) {
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(canceled)
		<-release // Simulate completing commit/cleanup after the client has left.
	}))
	defer server.Close()
	defer unblock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		request, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		response, _ := http.DefaultClient.Do(request)
		if response != nil {
			_ = response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("test request did not enter handler")
	}
	cancel()
	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not observe canceled request")
	}
	finished := make(chan struct{})
	go func() { finishCombatLoadHTTP(server, func() { <-clientDone }); close(finished) }()
	returnedEarly := false
	select {
	case <-finished:
		returnedEarly = true
	case <-time.After(50 * time.Millisecond):
	}
	unblock()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("server handler cleanup did not finish")
	}
	if returnedEarly {
		t.Fatal("final inspection can run before timed-out handler completes")
	}
}

func TestCombatLoadTimeoutDoesNotHideMissingAcknowledgedCommand(t *testing.T) {
	stats := &loadHTTPStats{Completed: 101, TransportErrors: 1, Statuses: map[int]int{200: 100}}
	if err := checkLoadCommandCounts(99, stats); err == nil {
		t.Fatal("one ambiguous response concealed a missing acknowledged command")
	}
	for _, count := range []int{100, 101} {
		if err := checkLoadCommandCounts(count, stats); err != nil {
			t.Fatal("valid ambiguous outcome rejected", count, err)
		}
	}
	stats = &loadHTTPStats{Completed: 3, Statuses: map[int]int{200: 1, 409: 2}}
	if err := checkLoadCommandCounts(2, stats); err == nil {
		t.Fatal("rejected command accepted as an ambiguous commit")
	}
}

func TestCombatLoadIntegrityDetectsLostResourceAndNPCUpdates(t *testing.T) {
	for _, resource := range []string{"energy", "energy_remainder", "health", "npc"} {
		t.Run(resource, func(t *testing.T) {
			f := newExecutionFixture(t)
			f.due(t)
			if _, err := f.r.ProcessDue(f.ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := f.cast("load:review"); err != nil {
				t.Fatal(err)
			}
			result := &loadResult{Battles: 1}
			if err := checkCombatLoadIntegrity(f.ctx, f.s.pool, result); err != nil {
				t.Fatal("valid fixture rejected", err)
			}
			switch resource {
			case "energy":
				mustExec(t, f.ctx, f.s, `UPDATE characters SET mp_current=mp WHERE id=$1`, f.in.CharacterID)
			case "energy_remainder":
				mustExec(t, f.ctx, f.s, `UPDATE characters SET mp_regen_remainder=mp_regen_remainder+1 WHERE id=$1`, f.in.CharacterID)
			case "health":
				mustExec(t, f.ctx, f.s, `UPDATE characters SET hp_current=hp WHERE id=$1`, f.in.CharacterID)
			case "npc":
				mustExec(t, f.ctx, f.s, `UPDATE battles SET npc_hp_current=npc_hp WHERE id=$1`, f.battle.ID)
			}
			if err := checkCombatLoadIntegrity(f.ctx, f.s.pool, result); err == nil {
				t.Fatal("lost resource update passed load integrity checks", resource)
			}
		})
	}
}
