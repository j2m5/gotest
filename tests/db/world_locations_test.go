package db_test

import (
	"gotest/internal/seed"
	"path/filepath"
	"testing"
)

const locationMigration = "20261104000200_character_locations"

func TestCharacterLocationDefaultsAndConstraints(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var ids []int
	var worlds []string
	if err := s.pool.QueryRow(ctx, `SELECT array_agg(c.id ORDER BY c.id),array_agg(p.code ORDER BY c.id) FROM characters c JOIN locations l ON l.id=c.location_id JOIN planets p ON p.id=l.planet_id`).Scan(&ids, &worlds); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || worlds[0] != "dromund_kaas" || worlds[1] != "coruscant" || worlds[2] != "coruscant" {
		t.Fatalf("starting worlds: %v %v", ids, worlds)
	}
	for _, sql := range []string{
		`UPDATE characters SET location_id=NULL WHERE id=1`,
		`UPDATE characters SET location_id=2147483647 WHERE id=1`,
		`UPDATE locations SET recommended_level=0`,
		`UPDATE locations SET position=11`,
		`INSERT INTO locations(planet_id,code,name,position,recommended_level) SELECT planet_id,code,'duplicate',2,1 FROM locations LIMIT 1`,
		`DELETE FROM locations WHERE id=(SELECT location_id FROM characters WHERE id=1)`,
	} {
		if _, err := s.pool.Exec(ctx, sql); err == nil {
			t.Fatalf("invalid location state accepted: %s", sql)
		}
	}
	var before, after string
	snapshot := `SELECT jsonb_agg(to_jsonb(l) ORDER BY id)::text FROM locations l`
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
		t.Fatalf("seed altered locations: %v", err)
	}
}

func TestCharacterLocationMigrationBackfillAndRollback(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, locationMigration)
	before := func() string {
		var value string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(c)-'location_id'-'updated_at' ORDER BY id)::text FROM characters c`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, locationMigration+"."+direction+".sql")
		var after string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(c)-'location_id'-'updated_at' ORDER BY id)::text FROM characters c`).Scan(&after); err != nil || after != before {
			t.Fatalf("%s changed characters: %v", direction, err)
		}
		if direction == "up" {
			var positioned int
			if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM characters WHERE location_id IS NOT NULL`).Scan(&positioned); err != nil || positioned != 3 {
				t.Fatalf("backfill %d: %v", positioned, err)
			}
		}
	}
}
