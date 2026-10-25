package db_test

import (
	"context"
	"strings"
	"testing"
)

func characterStatsConstraintCount(t *testing.T, ctx context.Context, s *testStorage) int {
	t.Helper()
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM pg_constraint WHERE connamespace=current_schema()::regnamespace
		AND convalidated AND conname IN ('characters_class_role_fkey','characters_level_fkey',
		'characters_experience_nonnegative_check','characters_hp_bounds_check','characters_mp_bounds_check',
		'characters_attributes_nonnegative_check','levels_level_positive_check','levels_experience_nonnegative_check')`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestCharacterStatsMigrationPreservesRowsAndRollback(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, characterStatsMigration)
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1;
		UPDATE characters SET role='tank',level=50,experience=154200,hp_current=0,mp=0,mp_current=0 WHERE id=1;
		INSERT INTO character_abilities(character_id,ability_id) SELECT 1,ability_id FROM class_abilities WHERE class_id=(SELECT class_id FROM characters WHERE id=1)`)
	before := characterStatsSnapshot(t, ctx, s)
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, characterStatsMigration+"."+direction+".sql")
		if characterStatsSnapshot(t, ctx, s) != before {
			t.Fatalf("%s changed account, character, abilities or catalogs", direction)
		}
		wantConstraints, wantIndexes := 8, 2
		if direction == "down" {
			wantConstraints, wantIndexes = 0, 0
		}
		if count := characterStatsConstraintCount(t, ctx, s); count != wantConstraints {
			t.Fatalf("%s validated constraints=%d; want %d", direction, count, wantConstraints)
		}
		var indexes int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema()
			AND indexname IN ('characters_class_role_idx','characters_level_idx')`).Scan(&indexes); err != nil || indexes != wantIndexes {
			t.Fatalf("%s indexes=%d; want %d; err=%v", direction, indexes, wantIndexes, err)
		}
		// Old ownership and faction checks must survive either direction.
		_, err := s.pool.Exec(ctx, `UPDATE users SET character_id=3 WHERE id=1`)
		requireConstraint(t, err, "23503", "users_active_character_owner_fkey")
		_, err = s.pool.Exec(ctx, `UPDATE characters SET faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE id=1`)
		requireConstraint(t, err, "23503", "characters_class_faction_fkey")
		if direction == "down" {
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, updateErr := tx.Exec(ctx, `UPDATE characters SET role='healer',level=99,experience=-1 WHERE id=1`)
			rollbackErr := tx.Rollback(ctx)
			if updateErr != nil || rollbackErr != nil {
				t.Fatalf("down did not restore old write behavior: %v %v", updateErr, rollbackErr)
			}
		}
	}
}

func TestCharacterStatsMigrationRejectsInvalidLegacyData(t *testing.T) {
	for _, tc := range []struct{ name, sql, diagnostic string }{
		{"unsupported role", `UPDATE characters SET role='healer' WHERE id=1`, "class roles"},
		{"missing role", `DELETE FROM class_roles WHERE class_id=(SELECT class_id FROM characters WHERE id=1) AND role='dd'`, "class roles"},
		{"missing level", `UPDATE characters SET level=99 WHERE id=1`, "character levels"},
		{"zero catalog level", `INSERT INTO levels(level,experience_required) VALUES(0,0)`, "levels catalog"},
		{"negative catalog experience", `UPDATE levels SET experience_required=-1 WHERE level=50`, "levels catalog"},
		{"negative experience", `UPDATE characters SET experience=-1 WHERE id=1`, "numeric values"},
		{"zero maximum HP", `UPDATE characters SET hp=0,hp_current=0 WHERE id=1`, "numeric values"},
		{"excess current HP", `UPDATE characters SET hp_current=hp+1 WHERE id=1`, "numeric values"},
		{"negative maximum MP", `UPDATE characters SET mp=-1,mp_current=0 WHERE id=1`, "numeric values"},
		{"excess current MP", `UPDATE characters SET mp_current=mp+1 WHERE id=1`, "numeric values"},
		{"negative attribute", `UPDATE characters SET absorption=-1 WHERE id=1`, "numeric values"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, ctx := characterTestStorageBefore(t, characterStatsMigration)
			mustExec(t, ctx, s, tc.sql)
			before := characterStatsSnapshot(t, ctx, s)
			conn, err := s.pool.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Release()
			_, err = conn.Exec(ctx, migrationSQL(t, characterStatsMigration+".up.sql"))
			if err == nil || !strings.Contains(err.Error(), "character stats:") || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("missing actionable legacy-data diagnostic: %v", err)
			}
			if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
				t.Fatal(err)
			}
			if characterStatsConstraintCount(t, ctx, s) != 0 || characterStatsSnapshot(t, ctx, s) != before {
				t.Fatal("failed migration repaired/deleted rows or left partial constraints")
			}
		})
	}
}

func TestCharacterStatsMigrationRollsBackLateFailure(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, characterStatsMigration)
	mustExec(t, ctx, s, `CREATE INDEX characters_level_idx ON characters(name)`)
	before := characterStatsSnapshot(t, ctx, s)
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, migrationSQL(t, characterStatsMigration+".up.sql"))
	if err == nil || !strings.Contains(err.Error(), "characters_level_idx") {
		t.Fatalf("expected late index collision: %v", err)
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	var newIndexes int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema()
		AND indexname='characters_class_role_idx'`).Scan(&newIndexes); err != nil || newIndexes != 0 {
		t.Fatalf("partial migration index remains: %d %v", newIndexes, err)
	}
	if characterStatsConstraintCount(t, ctx, s) != 0 || characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("late migration failure changed data or left constraints")
	}
}
