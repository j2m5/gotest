package db_test

import (
	"path/filepath"
	"testing"

	"gotest/internal/seed"
)

const npcMigration = "20261104000300_npc_catalog"

func TestNPCCatalogContentAndIntegrity(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var templates, placements, spawns, areas, minLevel, maxLevel int
	err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM npc_templates),(SELECT count(*) FROM location_npcs),(SELECT count(*) FROM npc_spawns),count(DISTINCT location_id),min(level),max(level) FROM location_npcs`).Scan(&templates, &placements, &spawns, &areas, &minLevel, &maxLevel)
	if err != nil {
		t.Fatal(err)
	}
	if templates != 415 || placements != 3000 || spawns != 4000 || areas != 1000 || minLevel != 1 || maxLevel != 50 {
		t.Fatalf("catalog counts %d/%d/%d/%d levels%d..%d", templates, placements, spawns, areas, minLevel, maxLevel)
	}
	var bad int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM (SELECT n.location_id,count(*) AS total,count(*) FILTER(WHERE n.rank='normal') AS normal,count(*) FILTER(WHERE n.rank='strong') AS strong,count(*) FILTER(WHERE n.rank='elite') AS elite FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id GROUP BY n.location_id) counts WHERE total<>4 OR normal<>2 OR strong<>1 OR elite<>1`).Scan(&bad); err != nil || bad != 0 {
		t.Fatalf("uneven population: %d %v", bad, err)
	}
	for _, sql := range []string{
		`UPDATE npc_templates SET kind='unknown'`,
		`UPDATE location_npcs SET level=0`,
		`UPDATE location_npcs SET rank='boss'`,
		`UPDATE location_npcs SET position=0`,
		`UPDATE location_npcs SET location_id=2147483647`,
		`UPDATE location_npcs SET npc_template_id=2147483647`,
		`UPDATE npc_spawns SET location_npc_id=2147483647`,
		`UPDATE npc_spawns SET slot=0`,
		`INSERT INTO npc_spawns(location_npc_id,slot) SELECT location_npc_id,slot FROM npc_spawns LIMIT 1`,
		`DELETE FROM npc_templates WHERE id=(SELECT npc_template_id FROM location_npcs LIMIT 1)`,
	} {
		if _, err = s.pool.Exec(ctx, sql); err == nil {
			t.Fatalf("invalid catalog accepted: %s", sql)
		}
	}
	mustExec(t, ctx, s, `UPDATE npc_templates SET description='User-authored enemy' WHERE id=(SELECT min(id) FROM npc_templates); UPDATE location_npcs SET level=2,rank='strong' WHERE id=(SELECT min(id) FROM location_npcs)`)
	snapshot := `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM npc_templates t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM location_npcs t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM npc_spawns t))::text`
	var before, after string
	if err = s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err = seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
		t.Fatal(err)
	}
	if err = s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || after != before {
		t.Fatalf("reseed rewrote NPC data: %v", err)
	}
}

func TestNPCMigrationRollbackPreservesWorld(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, npcMigration)
	snapshot := `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM characters c),(SELECT jsonb_agg(to_jsonb(l) ORDER BY id) FROM locations l),(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM planets p))::text`
	var before, after string
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, npcMigration+"."+direction+".sql")
		if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || after != before {
			t.Fatalf("%s changed world or characters: %v", direction, err)
		}
	}
}
