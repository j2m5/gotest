package db_test

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/maintenance"
	"gotest/internal/middleware"
	"gotest/tests/testenv"
)

func TestCombatRetentionBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
	f := cloneFinishedLog(t, wonLogFixture(t), 40)
	events := f.view(t).Events
	reward := events[len(events)-1]
	if _, err := maintenance.Run(f.ctx, f.s.pool, maintenance.Options{BatchSize: 2, BattleLogs: true}); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("d", 64)
	repositoryActor(t, f.ctx, f.s, f.actor.UserID, auth.TokenDigest(token))
	h, _ := characterFlowRouter(t, f.s, &flowMail{})
	server := httptest.NewServer(h)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profile := t.TempDir()
	output, err := runCombatBrowserDriver(ctx, browser, profile,
		filepath.Join("..", "frontend", "combat-retention.browser.mjs"), server.URL, browser, profile,
		strconv.FormatInt(f.battle.ID, 10), middleware.SessionCookie, token,
		strconv.FormatInt(reward.Experience, 10), strconv.FormatInt(reward.Credits, 10))
	if err != nil {
		t.Fatalf("archived combat browser: %v\n%s", err, output)
	}
	t.Log(string(output))
}
