package db_test

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gotest/tests/testenv"
)

func TestCombatReliabilityBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
	s, setup, _, _, in := combatFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("browser-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, setup, s, `UPDATE users SET password=$1`, string(hash))
	mustExec(t, setup, s, `INSERT INTO character_abilities(character_id,ability_id)
 SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id WHERE c.id IN(2,3)`)
	mustExec(t, setup, s, `UPDATE characters SET hp=1000,hp_current=1000,power=0,mastery=0 WHERE id IN(2,3);
 UPDATE abilities SET damage=5,mp_cost=10,cooldown_ms=3000`)
	var planetID int
	if err = s.pool.QueryRow(setup, `SELECT planet_id FROM locations WHERE id=$1`, in.LocationID).Scan(&planetID); err != nil {
		t.Fatal(err)
	}
	h, _ := characterFlowRouter(t, s, &flowMail{})
	server := httptest.NewServer(h)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	profile := t.TempDir()
	output, err := runCombatBrowserDriver(ctx, browser, profile, filepath.Join("..", "frontend", "combat-reliability.browser.mjs"), server.URL, browser, profile, strconv.Itoa(planetID), strconv.Itoa(in.LocationID))
	if err != nil {
		t.Fatalf("combat reliability browser: %v\n%s", err, output)
	}
	t.Log(string(output))
	var active, retreated, rewards int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status='active'),count(*) FILTER(WHERE status='retreated'),
 (SELECT count(*) FROM wallet_entries WHERE source='pve') FROM battles`).Scan(&active, &retreated, &rewards); err != nil {
		t.Fatal(err)
	}
	if active != 0 || retreated != 2 || rewards != 0 {
		t.Fatalf("unexpected final state: active%d retreated%d rewards%d", active, retreated, rewards)
	}
}
