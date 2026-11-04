package db_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest/internal/middleware"
	"gotest/internal/seed"
)

const planetMigration = "20261104000100_planets"

func TestWorldCatalogSeedAndCoreIntegrity(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var total, republicCore, empireCore, outer, neutral, republic, empire int
	err := s.pool.QueryRow(ctx, `SELECT count(*),
	 count(*) FILTER(WHERE cf.code='galactic_republic'), count(*) FILTER(WHERE cf.code='sith_empire'),
	 count(*) FILTER(WHERE p.core_faction_id IS NULL), count(*) FILTER(WHERE p.controlling_faction_id IS NULL),
	 count(*) FILTER(WHERE f.code='galactic_republic'), count(*) FILTER(WHERE f.code='sith_empire')
	 FROM planets p LEFT JOIN factions cf ON cf.id=p.core_faction_id LEFT JOIN factions f ON f.id=p.controlling_faction_id`).Scan(
		&total, &republicCore, &empireCore, &outer, &neutral, &republic, &empire)
	if err != nil || total != 100 || republicCore != 15 || empireCore != 15 || outer != 70 || neutral != 50 || republic != 25 || empire != 25 {
		t.Fatalf("planet distribution: %d/%d/%d/%d, control %d/%d/%d: %v", total, republicCore, empireCore, outer, neutral, republic, empire, err)
	}
	for _, sql := range []string{
		`UPDATE planets SET controlling_faction_id=NULL WHERE code='coruscant'`,
		`UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire') WHERE code='coruscant'`,
		`UPDATE planets SET core_faction_id=NULL WHERE code='coruscant'`,
		`UPDATE planets SET (core_faction_id,controlling_faction_id)=(SELECT id,id FROM factions WHERE code='sith_empire') WHERE code='coruscant'`,
		`UPDATE planets SET controlling_faction_id=2147483647 WHERE code='tatooine'`,
		`UPDATE planets SET code='coruscant' WHERE code='tatooine'`,
		`UPDATE planets SET name=' ' WHERE code='tatooine'`,
	} {
		if _, err := s.pool.Exec(ctx, sql); err == nil {
			t.Fatalf("accepted invalid planet state: %s", sql)
		}
	}
	mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire'),description='Edited world' WHERE code='tatooine'`)
	var before, after string
	snapshot := `SELECT jsonb_agg(to_jsonb(p) ORDER BY id)::text FROM planets p`
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
		t.Fatalf("seed reset world state: %v", err)
	}
}

func TestWorldMigrationMetadataAndRollback(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, locationMigration)
	var created, updated time.Time
	if err := s.pool.QueryRow(ctx, `SELECT created_at,updated_at FROM planets WHERE code='tatooine'`).Scan(&created, &updated); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE planets SET description='New description' WHERE code='tatooine'`)
	var original time.Time
	var changed time.Time
	if err := s.pool.QueryRow(ctx, `SELECT created_at,updated_at FROM planets WHERE code='tatooine'`).Scan(&original, &changed); err != nil || !original.Equal(created) || !changed.After(updated) {
		t.Fatalf("planet timestamps: created=%v updated=%v: %v", original, changed, err)
	}
	executeMigration(t, ctx, s, planetMigration+".down.sql")
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM characters`).Scan(&count); err != nil || count != 3 {
		t.Fatalf("rollback changed characters: %d %v", count, err)
	}
	executeMigration(t, ctx, s, planetMigration+".up.sql")
	planetSeed, err := os.ReadFile(filepath.Join("..", "..", "seeds", "07_planets.sql"))
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, string(planetSeed))
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM planets`).Scan(&count); err != nil || count != 100 {
		t.Fatalf("up/down/up: %d %v", count, err)
	}
}

func TestWorldCatalogThroughHTTP(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, locationMigration)
	h, _ := characterFlowRouter(t, s, &flowMail{})
	flowRequest(t, h, "GET", "/api/world/planets", "", nil, 401)
	raw := strings.Repeat("e", 64)
	hash := sha256.Sum256([]byte(raw))
	if err := createStoredSession(t, s, ctx, 1, hex.EncodeToString(hash[:]), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: raw}
	w := flowRequest(t, h, "GET", "/api/world/planets", "", cookie, 200)
	var result struct {
		Planets []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			NameEN string `json:"name_en"`
			Core   *struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"core_faction"`
			Control *struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"controlling_faction"`
		} `json:"planets"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Planets) != 100 {
		t.Fatalf("catalog: %v %s", err, w.Body)
	}
	found := false
	for _, p := range result.Planets {
		if p.Code == "coruscant" {
			found = true
			if p.Name != "Корускант" || p.NameEN != "Coruscant" || p.Core == nil || p.Control == nil || p.Core.Code != "galactic_republic" || p.Control.Name != "Республика" {
				t.Fatalf("core display: %+v", p)
			}
		}
	}
	if !found {
		t.Fatal("Coruscant missing")
	}
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("catalog selected a character")
	}
	flowRequest(t, h, "POST", "/api/world/planets", `{}`, cookie, 405)
	mustExec(t, ctx, s, `DELETE FROM planets`)
	w = flowRequest(t, h, "GET", "/api/world/planets", "", cookie, 200)
	if strings.TrimSpace(w.Body.String()) != `{"planets":[]}` {
		t.Fatalf("empty catalog: %s", w.Body)
	}
}
