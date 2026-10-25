package seed_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func migrationPath(direction string) string {
	return filepath.Join("..", "..", "migrations", codeMigration+"."+direction+".sql")
}

func TestSeedMigrationPreservesLegacyReferences(t *testing.T) {
	s := newSeedDB(t, true)
	s.execFile(t, filepath.Join("testdata", "legacy.sql"))
	before := s.snapshot(t)
	s.execFile(t, migrationPath("up"))
	var first, second, custom string
	err := s.pool.QueryRow(s.ctx, `SELECT
		(SELECT code FROM abilities WHERE id=201),
		(SELECT code FROM abilities WHERE id=202),
		(SELECT code FROM abilities WHERE id=299)`).Scan(&first, &second, &custom)
	if err != nil || first != "sith_inquisitor_basic_attack" || second != "jedi_consular_basic_attack" || custom != "legacy_ability_299" {
		t.Fatalf("ability mapping: %q, %q, %q; err=%v", first, second, custom, err)
	}
	s.execFile(t, migrationPath("down"))
	if s.snapshot(t) != before {
		t.Fatal("up/down changed existing catalog rows or their references")
	}
	s.execFile(t, migrationPath("up"))
	s.copySeeds(t, "seeds")
	if output, err := s.run(nil); err != nil {
		t.Fatalf("seed migrated catalog: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{3, 9, 9, 20, 8, 50})
	var classID, abilityID, damage int
	err = s.pool.QueryRow(s.ctx, `SELECT c.class_id,ca.ability_id,a.damage FROM characters c
		JOIN character_abilities ca ON ca.character_id=c.id JOIN abilities a ON a.id=ca.ability_id
		WHERE c.id=1`).Scan(&classID, &abilityID, &damage)
	if err != nil || classID != 101 || abilityID != 201 || damage != 77 {
		t.Fatalf("existing character/ability changed: class=%d ability=%d damage=%d err=%v", classID, abilityID, damage, err)
	}
}

func TestSeedMigrationRejectsAmbiguousLegacyRowsAtomically(t *testing.T) {
	for _, tt := range []struct{ name, sql string }{
		{"factions", "INSERT INTO factions(name,alias) VALUES ('Sith Empire','one'),('Sith Empire','two')"},
		{"classes", "INSERT INTO factions(name,alias) VALUES ('Sith Empire','Empire'); INSERT INTO classes(faction_id,name) VALUES (1,'Sith Warrior'),(1,'Sith Warrior')"},
		{"abilities", "INSERT INTO abilities(name,icon) VALUES ('Удар мечем','images/abilities/1562632956_1963234752.png'),('Удар мечем','images/abilities/1562632956_1963234752.png')"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newSeedDB(t, true)
			s.exec(t, tt.sql)
			before := s.snapshot(t)
			conn, err := s.pool.Acquire(s.ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Release()
			_, err = conn.Exec(s.ctx, readFile(t, migrationPath("up")))
			if err == nil || !strings.Contains(err.Error(), "duplicate") {
				t.Fatalf("ambiguous legacy rows must stop migration with a diagnostic: %v", err)
			}
			if _, err := conn.Exec(s.ctx, "ROLLBACK"); err != nil {
				t.Fatal(err)
			}
			var codeColumns int
			if err := conn.QueryRow(s.ctx, `SELECT count(*) FROM information_schema.columns
				WHERE table_schema=current_schema() AND table_name IN ('factions','classes','abilities') AND column_name='code'`).Scan(&codeColumns); err != nil || codeColumns != 0 {
				t.Fatalf("failed migration left schema changes: %d, err=%v", codeColumns, err)
			}
			if s.snapshot(t) != before {
				t.Fatal("failed migration changed or deleted existing data")
			}
		})
	}
}
