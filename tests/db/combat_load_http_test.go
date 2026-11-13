package db_test

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"gotest/internal/middleware"
)

type loadJob struct {
	actor  loadActor
	kind   string
	offset time.Duration
	key    string
}

func combatLoadJobs(actors []loadActor, duration time.Duration) []loadJob {
	var jobs []loadJob
	for i, actor := range actors {
		for at := time.Duration(i) * (2 * time.Second) / time.Duration(len(actors)); at < duration; at += 2 * time.Second {
			for _, kind := range []string{"current", "detail", "abilities"} {
				jobs = append(jobs, loadJob{actor: actor, kind: kind, offset: at})
			}
		}
		for at := time.Duration(i)*(3*time.Second)/time.Duration(len(actors)) + 100*time.Millisecond; at < duration; at += 3 * time.Second {
			jobs = append(jobs, loadJob{actor: actor, kind: "use", offset: at, key: fmt.Sprintf("load:%d:%d", actor.userID, at)})
		}
	}
	slices.SortStableFunc(jobs, func(a, b loadJob) int { return cmp.Compare(a.offset, b.offset) })
	return jobs
}

func waitCombatLoadUntil(ctx context.Context, at time.Time) bool {
	timer := time.NewTimer(time.Until(at))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Fixed arrivals keep slow responses from silently lowering the offered load.
// The bounded queue counts overflow; elapsed latency includes time waiting here.
func startCombatLoadHTTP(ctx context.Context, baseURL string, actors []loadActor, start time.Time, duration time.Duration) (*loadHTTPMetrics, func()) {
	metrics := &loadHTTPMetrics{}
	jobs := combatLoadJobs(actors, duration)
	queue := make(chan loadJob, 256)
	transport := &http.Transport{MaxIdleConns: 64, MaxIdleConnsPerHost: 64, MaxConnsPerHost: 64}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	var clients sync.WaitGroup
	for range 64 {
		clients.Go(func() {
			for job := range queue {
				scheduled, started := start.Add(job.offset), time.Now()
				code, err := performCombatLoadRequest(ctx, client, baseURL, job)
				metrics.record(job.kind, job.key, scheduled, started, time.Now(), code, err != nil, false)
			}
		})
	}
	produced := make(chan struct{})
	go func() {
		defer close(produced)
		defer close(queue)
		for _, job := range jobs {
			at := start.Add(job.offset)
			if !waitCombatLoadUntil(ctx, at) {
				metrics.record(job.kind, job.key, at, at, at, 0, false, true)
				continue
			}
			select {
			case queue <- job:
			default:
				metrics.record(job.kind, job.key, at, at, at, 0, false, true)
			}
		}
	}()
	return metrics, func() { <-produced; clients.Wait(); transport.CloseIdleConnections() }
}

func performCombatLoadRequest(ctx context.Context, client *http.Client, baseURL string, job loadJob) (int, error) {
	actor := job.actor
	path, method, body := "", "GET", ""
	switch job.kind {
	case "current":
		path = fmt.Sprintf("/api/battles/current?character_id=%d", actor.characterID)
	case "detail":
		path = fmt.Sprintf("/api/battles/detail?battle_id=%d&character_id=%d", actor.battleID, actor.characterID)
	case "abilities":
		path = fmt.Sprintf("/api/abilities?character_id=%d", actor.characterID)
	case "use":
		path, method = "/api/battles/use", "POST"
		body = fmt.Sprintf(`{"battle_id":"%d","character_id":%d,"ability_id":%d,"request_key":%q}`, actor.battleID, actor.characterID, actor.abilityID, job.key)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, strings.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: actor.token})
	req.Header.Set("X-Account-ID", strconv.Itoa(actor.userID))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	return response.StatusCode, err
}
