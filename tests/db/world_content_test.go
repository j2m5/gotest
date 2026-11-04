package db_test

import (
	"path/filepath"
	"testing"

	"gotest/internal/db"
	"gotest/internal/seed"
	"gotest/internal/wallet"
)

func TestAdditionalWorldContentPreservesExistingState(t *testing.T) {
	t.Run("capital_batch", func(t *testing.T) { assertWorldContentExpansion(t, "09_npcs.sql", true, 22, 80) })
	t.Run("four_world_batch", func(t *testing.T) { assertWorldContentExpansion(t, "09_npcs.sql", false, 40, 160) })
	t.Run("first_frontier_batch", func(t *testing.T) { assertWorldContentExpansion(t, "10_frontier_worlds.sql", false, 100, 400) })
	t.Run("second_frontier_batch", func(t *testing.T) { assertWorldContentExpansion(t, "11_frontier_worlds.sql", false, 160, 640) })
	t.Run("complete_frontier_batch", func(t *testing.T) { assertWorldContentExpansion(t, "12_frontier_worlds.sql", false, 240, 960) })
	t.Run("first_core_batch", func(t *testing.T) { assertWorldContentExpansion(t, "13_core_worlds.sql", false, 300, 1200) })
	t.Run("second_core_batch", func(t *testing.T) { assertWorldContentExpansion(t, "14_core_worlds.sql", false, 360, 1440) })
	t.Run("third_core_batch", func(t *testing.T) { assertWorldContentExpansion(t, "15_core_worlds.sql", false, 420, 1680) })
	t.Run("complete_core_batch", func(t *testing.T) { assertWorldContentExpansion(t, "16_core_worlds.sql", false, 500, 2000) })
	t.Run("first_neutral_batch", func(t *testing.T) { assertWorldContentExpansion(t, "17_neutral_worlds.sql", false, 580, 2320) })
	t.Run("second_neutral_batch", func(t *testing.T) { assertWorldContentExpansion(t, "18_neutral_worlds.sql", false, 660, 2640) })
	t.Run("third_neutral_batch", func(t *testing.T) { assertWorldContentExpansion(t, "19_neutral_worlds.sql", false, 740, 2960) })
	t.Run("fourth_neutral_batch", func(t *testing.T) { assertWorldContentExpansion(t, "20_neutral_worlds.sql", false, 820, 3280) })
	t.Run("fifth_neutral_batch", func(t *testing.T) { assertWorldContentExpansion(t, "21_neutral_worlds.sql", false, 910, 3640) })
}

func assertWorldContentExpansion(t *testing.T, throughSeed string, capitalsOnly bool, oldLocations, oldNPCs int) {
	t.Helper()
	s, ctx := characterTestStorageThroughSeed(t, "", throughSeed)
	// Load only files belonging to the previous batch. The earliest capital-only
	// batch predates additions inside seed08/09, so reconstruct that one explicitly.
	if capitalsOnly {
		mustExec(t, ctx, s, `
		DELETE FROM npc_spawns WHERE location_npc_id IN (
			SELECT n.id FROM location_npcs n JOIN locations l ON l.id=n.location_id
			JOIN planets p ON p.id=l.planet_id WHERE p.code IN ('tython','korriban'));
		DELETE FROM location_npcs WHERE location_id IN (
			SELECT l.id FROM locations l JOIN planets p ON p.id=l.planet_id WHERE p.code IN ('tython','korriban'));
		DELETE FROM locations WHERE position>1 AND planet_id IN (SELECT id FROM planets WHERE code IN ('tython','korriban'));
		DELETE FROM npc_templates WHERE code LIKE 'tython\_%' ESCAPE '\' OR code LIKE 'korriban\_%' ESCAPE '\'`)
	}
	var previousLocations, previousNPCs int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM locations),(SELECT count(*) FROM npc_spawns)`).Scan(&previousLocations, &previousNPCs); err != nil || previousLocations != oldLocations || previousNPCs != oldNPCs {
		t.Fatalf("previous batch: locations=%d NPCs=%d err=%v", previousLocations, previousNPCs, err)
	}
	mustExec(t, ctx, s, `
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire') WHERE code='balmorra';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE code='tatooine';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire') WHERE code='makeb';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE code='rishi';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire') WHERE code='kintan';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE code='taanab';
		UPDATE planets SET controlling_faction_id=(SELECT id FROM factions WHERE code='sith_empire') WHERE code='mygeeto';
		UPDATE characters SET location_id=(SELECT l.id FROM locations l JOIN planets p ON p.id=l.planet_id WHERE p.code='korriban' AND l.position=1) WHERE id=1;
		UPDATE locations SET description='Edited entry area',recommended_level=2 WHERE id=(SELECT location_id FROM characters WHERE id=1);
		UPDATE npc_templates SET description='Edited enemy' WHERE code='tomb_raider';
		UPDATE location_npcs SET level=4,rank='strong' WHERE id=(SELECT min(id) FROM location_npcs)`)
	if _, err := db.NewWalletRepository(s.pool).Apply(ctx, walletOperation(1, wallet.Credits, 25, "before-content")); err != nil {
		t.Fatal(err)
	}
	// Remember exactly the existing rows, including timestamps and IDs. New rows
	// must be added without rewriting this content or any player state.
	mustExec(t, ctx, s, `
		CREATE TABLE prior_locations AS SELECT id FROM locations;
		CREATE TABLE prior_templates AS SELECT id FROM npc_templates;
		CREATE TABLE prior_placements AS SELECT id FROM location_npcs;
		CREATE TABLE prior_spawns AS SELECT id FROM npc_spawns`)
	snapshot := `SELECT jsonb_build_array(
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM planets t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM users t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM characters t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM wallets t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM wallet_entries t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM locations t WHERE id IN (SELECT id FROM prior_locations)),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM npc_templates t WHERE id IN (SELECT id FROM prior_templates)),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM location_npcs t WHERE id IN (SELECT id FROM prior_placements)),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM npc_spawns t WHERE id IN (SELECT id FROM prior_spawns)))::text`
	var before, after string
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
			t.Fatal(err)
		}
		if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || after != before {
			t.Fatalf("content expansion changed existing rows: %v", err)
		}
		var locations, instances int
		if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM locations),(SELECT count(*) FROM npc_spawns)`).Scan(&locations, &instances); err != nil || locations != 1000 || instances != 4000 {
			t.Fatalf("incomplete expansion: locations=%d NPCs=%d err=%v", locations, instances, err)
		}
	}
}
