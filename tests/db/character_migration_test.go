package db_test

import (
	"strings"
	"testing"
)

func TestCharacterMigrationUpDownPreservesData(t *testing.T) {
	s, ctx := characterTestStorage(t, false)
	mustExec(t, ctx, s, "UPDATE users SET character_id=1 WHERE id=1")
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, characterMigration+"."+direction+".sql")
		// The migration changes a result column from bigint to integer (or back).
		// Reconnect after DDL so pgx does not reuse the old prepared result type.
		s.pool.Reset()
		wantType := "integer"
		if direction == "down" {
			wantType = "bigint"
		}
		var columnType string
		if err := s.pool.QueryRow(ctx, `SELECT data_type FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='users' AND column_name='character_id'`).Scan(&columnType); err != nil || columnType != wantType {
			t.Fatalf("%s character_id type=%s, want=%s, err=%v", direction, columnType, wantType, err)
		}
		if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 1 {
			t.Fatalf("%s changed selected character", direction)
		}
		var users, characters int
		if err := s.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM characters)").Scan(&users, &characters); err != nil || users != 2 || characters != 3 {
			t.Fatalf("%s changed accounts/characters: %d/%d, err=%v", direction, users, characters, err)
		}
		// A missing class remains invalid even after reverting the composite FK.
		_, err := s.pool.Exec(ctx, "UPDATE characters SET class_id=999999 WHERE id=1")
		wantForeignKey(t, err)
	}
}

func TestCharacterMigrationRejectsInvalidLegacyDataAtomically(t *testing.T) {
	for _, tt := range []struct{ name, sql string }{
		{"faction", "UPDATE characters SET faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE id=1"},
		{"owner", "UPDATE users SET character_id=3 WHERE id=1"},
		{"missing", "UPDATE users SET character_id=999999 WHERE id=1"},
		{"out_of_range", "UPDATE users SET character_id=2147483648 WHERE id=1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := characterTestStorage(t, false)
			mustExec(t, ctx, s, tt.sql)
			conn, err := s.pool.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Release()
			_, err = conn.Exec(ctx, migrationSQL(t, characterMigration+".up.sql"))
			if err == nil || !strings.Contains(err.Error(), "character integrity:") {
				t.Fatalf("expected actionable validation error: %v", err)
			}
			if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
				t.Fatal(err)
			}
			var columnType string
			var newConstraints, users, characters int
			err = conn.QueryRow(ctx, `SELECT
				(SELECT data_type FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='users' AND column_name='character_id'),
				(SELECT count(*) FROM pg_constraint WHERE connamespace=current_schema()::regnamespace AND conname IN ('users_active_character_owner_fkey','characters_class_faction_fkey','characters_user_id_id_key','classes_id_faction_id_key')),
				(SELECT count(*) FROM users),(SELECT count(*) FROM characters)`).Scan(&columnType, &newConstraints, &users, &characters)
			if err != nil || columnType != "bigint" || newConstraints != 0 || users != 2 || characters != 3 {
				t.Fatalf("failed migration changed schema/data: type=%s constraints=%d users=%d characters=%d err=%v", columnType, newConstraints, users, characters, err)
			}
		})
	}
}
