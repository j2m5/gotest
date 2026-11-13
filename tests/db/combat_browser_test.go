package db_test

import (
	"context"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gotest/tests/testenv"
)

func TestCombatBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
	s, ctx, _, _, in := combatFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("browser-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET password=$1`, string(hash))
	mustExec(t, ctx, s, `INSERT INTO character_abilities(character_id,ability_id) SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id WHERE c.id=2`)
	mustExec(t, ctx, s, `UPDATE characters SET power=0,mastery=0 WHERE id=2; UPDATE abilities SET damage=20,mp_cost=10,cooldown_ms=1500`)
	var planetID int
	if err = s.pool.QueryRow(ctx, `SELECT planet_id FROM locations WHERE id=$1`, in.LocationID).Scan(&planetID); err != nil {
		t.Fatal(err)
	}
	h, _ := characterFlowRouter(t, s, &flowMail{})
	server := httptest.NewServer(h)
	defer server.Close()
	browserCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(browserCtx, "node", filepath.Join("..", "frontend", "combat.browser.mjs"), server.URL, browser, t.TempDir(), strconv.Itoa(planetID), strconv.Itoa(in.LocationID))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("combat browser: %v\n%s", err, output)
	}
	t.Log(string(output))
	var won, retreated, rewards int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status='won'),count(*) FILTER(WHERE status='retreated'),(SELECT count(*) FROM wallet_entries WHERE source='pve') FROM battles`).Scan(&won, &retreated, &rewards); err != nil {
		t.Fatal(err)
	}
	if won != 1 || retreated != 2 || rewards != 1 {
		t.Fatalf("unexpected outcomes: won%d retreated%d rewards%d", won, retreated, rewards)
	}
}
