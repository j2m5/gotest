package db_test

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"sync"
	"testing"
	"time"
)

type loadHTTPStats struct {
	Scheduled       int           `json:"scheduled"`
	Completed       int           `json:"completed"`
	Dropped         int           `json:"dropped"`
	TransportErrors int           `json:"transport_errors"`
	Statuses        map[int]int   `json:"statuses"`
	LatencyMS       loadQuantiles `json:"scheduled_to_response_ms"`
	ServiceMS       loadQuantiles `json:"request_to_response_ms"`
	latency         []time.Duration
	service         []time.Duration
}

type loadQuantiles struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
	Max float64 `json:"max"`
}

func loadPercentile(values []time.Duration, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := slices.Clone(values)
	slices.Sort(ordered)
	index := max(0, min(len(ordered)-1, int(math.Ceil(p*float64(len(ordered))))-1))
	return float64(ordered[index]) / float64(time.Millisecond)
}

func loadDistribution(values []time.Duration) loadQuantiles {
	return loadQuantiles{loadPercentile(values, .5), loadPercentile(values, .95), loadPercentile(values, .99), loadPercentile(values, 1)}
}

type loadHTTPMetrics struct {
	mu       sync.Mutex
	kinds    map[string]*loadHTTPStats
	outcomes map[string]loadCommandState
}

type loadCommandState string

const (
	loadAcknowledged loadCommandState = "acknowledged"
	loadAmbiguous    loadCommandState = "ambiguous"
	loadRejected     loadCommandState = "rejected"
	loadNotSent      loadCommandState = "not_sent"
)

type loadCommandEvidence struct {
	Acknowledged       int `json:"acknowledged"`
	Ambiguous          int `json:"ambiguous"`
	AmbiguousCommitted int `json:"ambiguous_committed"`
	Rejected           int `json:"rejected"`
	NotSent            int `json:"not_sent"`
}

func checkLoadCommandEvidence(outcomes map[string]loadCommandState, committed map[string]bool) (loadCommandEvidence, error) {
	var evidence loadCommandEvidence
	for key := range committed {
		if _, ok := outcomes[key]; !ok {
			return evidence, fmt.Errorf("unissued command committed: %s", key)
		}
	}
	for key, state := range outcomes {
		switch state {
		case loadAcknowledged:
			evidence.Acknowledged++
			if !committed[key] {
				return evidence, fmt.Errorf("acknowledged command missing: %s", key)
			}
		case loadAmbiguous:
			evidence.Ambiguous++
			if committed[key] {
				evidence.AmbiguousCommitted++
			}
		case loadRejected, loadNotSent:
			if state == loadRejected {
				evidence.Rejected++
			} else {
				evidence.NotSent++
			}
			if committed[key] {
				return evidence, fmt.Errorf("rejected or unsent command committed: %s", key)
			}
		default:
			return evidence, fmt.Errorf("unknown load command state %q", state)
		}
	}
	return evidence, nil
}

func (m *loadHTTPMetrics) record(kind, key string, scheduled, started, ended time.Time, code int, failed, dropped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.kinds == nil {
		m.kinds = make(map[string]*loadHTTPStats)
	}
	if kind == "use" {
		if m.outcomes == nil {
			m.outcomes = make(map[string]loadCommandState)
		}
		state := loadRejected
		switch {
		case dropped:
			state = loadNotSent
		case failed || code >= 500:
			state = loadAmbiguous
		case code == 200:
			state = loadAcknowledged
		}
		m.outcomes[key] = state
	}
	s := m.kinds[kind]
	if s == nil {
		s = &loadHTTPStats{Statuses: make(map[int]int)}
		m.kinds[kind] = s
	}
	s.Scheduled++
	if dropped {
		s.Dropped++
		return
	}
	s.Completed++
	if failed {
		s.TransportErrors++
	} else {
		s.Statuses[code]++
	}
	s.latency = append(s.latency, ended.Sub(scheduled))
	s.service = append(s.service, ended.Sub(started))
}

func (m *loadHTTPMetrics) commandResults() map[string]loadCommandState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Clone(m.outcomes)
}

func (m *loadHTTPMetrics) result() map[string]*loadHTTPStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.kinds {
		s.LatencyMS, s.ServiceMS = loadDistribution(s.latency), loadDistribution(s.service)
		s.latency, s.service = nil, nil
	}
	return m.kinds
}

func TestCombatLoadMetricsIncludeQueueDelayAndDroppedWork(t *testing.T) {
	if loadPercentile(nil, .95) != 0 || loadPercentile([]time.Duration{4 * time.Millisecond, time.Millisecond, 3 * time.Millisecond, 2 * time.Millisecond}, .5) != 2 {
		t.Fatal("incorrect percentile rank")
	}
	now := time.Now()
	var metrics loadHTTPMetrics
	metrics.record("use", "success", now, now.Add(time.Second), now.Add(1200*time.Millisecond), 200, false, false)
	metrics.record("use", "dropped", now, now, now, 0, false, true)
	metrics.record("use", "timeout", now, now, now.Add(2*time.Second), 0, true, false)
	s := metrics.result()["use"]
	if s.Scheduled != 3 || s.Completed != 2 || s.Dropped != 1 || s.TransportErrors != 1 || s.Statuses[200] != 1 || s.LatencyMS.P50 != 1200 || s.ServiceMS.P50 != 200 || s.LatencyMS.P95 != 2000 {
		t.Fatalf("incomplete load accounting: %+v", s)
	}
}

func TestCombatLoadReconcilesEachRequestKey(t *testing.T) {
	outcomes := map[string]loadCommandState{"first": loadAcknowledged, "second": loadAmbiguous, "third": loadRejected, "fourth": loadNotSent}
	if _, err := checkLoadCommandEvidence(outcomes, map[string]bool{"second": true}); err == nil {
		t.Fatal("ambiguous commit replaced acknowledged command")
	}
	for _, unexpected := range []string{"third", "fourth", "unknown"} {
		if _, err := checkLoadCommandEvidence(outcomes, map[string]bool{"first": true, unexpected: true}); err == nil {
			t.Fatal("unaccepted key committed", unexpected)
		}
	}
	evidence, err := checkLoadCommandEvidence(outcomes, map[string]bool{"first": true, "second": true})
	if err != nil || evidence.Acknowledged != 1 || evidence.Ambiguous != 1 || evidence.AmbiguousCommitted != 1 || evidence.Rejected != 1 || evidence.NotSent != 1 {
		t.Fatal(evidence, err)
	}
}
