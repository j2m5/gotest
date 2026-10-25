package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gotest/internal/character"
	"gotest/internal/db"
)

const characterStatsMigration = "20261025000700_character_stats_integrity"

func requireConstraint(t *testing.T, err error, code, name string) {
	t.Helper()
	var postgres *pgconn.PgError
	if !errors.As(err, &postgres) || postgres.Code != code || postgres.ConstraintName != name {
		t.Fatalf("want %s from %s; got %v", code, name, err)
	}
}

func characterStatsSnapshot(t *testing.T, ctx context.Context, s *testStorage) string {
	t.Helper()
	var snapshot string
	if err := s.pool.QueryRow(ctx, `SELECT jsonb_build_array(
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM users t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM characters t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY class_id,role) FROM class_roles t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY level) FROM levels t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY character_id,ability_id) FROM character_abilities t)
	)::text`).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestCharacterRoleAndLevelReferences(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	before := characterStatsSnapshot(t, ctx, s)
	for _, tc := range []struct{ name, sql, constraint string }{
		{"insert unsupported role", `INSERT INTO characters(user_id,faction_id,class_id,name,role) SELECT 1,faction_id,id,'Invalid role','healer' FROM classes WHERE code='sith_warrior'`, "characters_class_role_fkey"},
		{"update unsupported role", `UPDATE characters SET role='healer' WHERE id=1`, "characters_class_role_fkey"},
		{"empty role", `UPDATE characters SET role='' WHERE id=1`, "characters_class_role_fkey"},
		{"remove used role", `DELETE FROM class_roles WHERE class_id=(SELECT id FROM classes WHERE code='sith_warrior') AND role='dd'`, "characters_class_role_fkey"},
		{"rename used role", `UPDATE class_roles SET role='renamed' WHERE class_id=(SELECT id FROM classes WHERE code='sith_warrior') AND role='dd'`, "characters_class_role_fkey"},
		{"insert unknown level", `INSERT INTO characters(user_id,faction_id,class_id,name,level) SELECT 1,faction_id,id,'Invalid level',99 FROM classes WHERE code='sith_warrior'`, "characters_level_fkey"},
		{"update unknown level", `UPDATE characters SET level=99 WHERE id=1`, "characters_level_fkey"},
		{"zero character level", `UPDATE characters SET level=0 WHERE id=1`, "characters_level_fkey"},
		{"remove used level", `DELETE FROM levels WHERE level=1`, "characters_level_fkey"},
		{"renumber used level", `UPDATE levels SET level=51 WHERE level=1`, "characters_level_fkey"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.pool.Exec(ctx, tc.sql)
			requireConstraint(t, err, "23503", tc.constraint)
			if characterStatsSnapshot(t, ctx, s) != before {
				t.Fatal("rejected reference change modified stored rows")
			}
		})
	}
	// Changing the class also rechecks the currently selected role.
	mustExec(t, ctx, s, `UPDATE characters SET role='tank' WHERE id=1`)
	_, err := s.pool.Exec(ctx, `UPDATE characters SET (class_id,faction_id)=(SELECT id,faction_id FROM classes WHERE code='imperial_agent') WHERE id=1`)
	requireConstraint(t, err, "23503", "characters_class_role_fkey")
}

func TestCharacterStatBoundsOnInsertAndUpdate(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, tc := range []struct {
		column     string
		value      int
		constraint string
	}{
		{"experience", -1, "characters_experience_nonnegative_check"},
		{"hp", 0, "characters_hp_bounds_check"}, {"hp", -1, "characters_hp_bounds_check"},
		{"hp_current", -1, "characters_hp_bounds_check"}, {"hp_current", 101, "characters_hp_bounds_check"},
		{"mp", -1, "characters_mp_bounds_check"},
		{"mp_current", -1, "characters_mp_bounds_check"}, {"mp_current", 101, "characters_mp_bounds_check"},
		{"mastery", -1, "characters_attributes_nonnegative_check"},
		{"endurance", -1, "characters_attributes_nonnegative_check"},
		{"power", -1, "characters_attributes_nonnegative_check"},
		{"accuracy", -1, "characters_attributes_nonnegative_check"},
		{"alacrity", -1, "characters_attributes_nonnegative_check"},
		{"critical", -1, "characters_attributes_nonnegative_check"},
		{"defense", -1, "characters_attributes_nonnegative_check"},
		{"shield", -1, "characters_attributes_nonnegative_check"},
		{"absorption", -1, "characters_attributes_nonnegative_check"},
	} {
		t.Run(tc.column, func(t *testing.T) {
			_, err := s.pool.Exec(ctx, "UPDATE characters SET "+tc.column+"=$1 WHERE id=1", tc.value)
			requireConstraint(t, err, "23514", tc.constraint)
			_, err = s.pool.Exec(ctx, "INSERT INTO characters(user_id,faction_id,class_id,name,"+tc.column+") SELECT 1,faction_id,id,'Invalid numeric value',$1 FROM classes WHERE code='sith_warrior'", tc.value)
			requireConstraint(t, err, "23514", tc.constraint)
		})
	}
	for _, tc := range []struct{ sql, constraint string }{
		{`UPDATE characters SET hp=99 WHERE id=1`, "characters_hp_bounds_check"},
		{`UPDATE characters SET mp=99 WHERE id=1`, "characters_mp_bounds_check"},
	} {
		_, err := s.pool.Exec(ctx, tc.sql)
		requireConstraint(t, err, "23514", tc.constraint)
	}
}

func TestCharacterStatsAllowValidBoundariesAndCatalogGrowth(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE characters SET hp=1,hp_current=0,mp=0,mp_current=0,experience=0,
		mastery=0,endurance=0,power=0,accuracy=0,alacrity=0,critical=0,defense=0,shield=0,absorption=0 WHERE id=1;
		INSERT INTO levels(level,experience_required) VALUES(51,200000);
		UPDATE characters SET level=51,hp=2147483647,hp_current=2147483647,mp=2147483647,mp_current=2147483647,
		experience=2147483647,mastery=2147483647,endurance=2147483647,power=2147483647,accuracy=2147483647,
		alacrity=2147483647,critical=2147483647,defense=2147483647,shield=2147483647,absorption=2147483647 WHERE id=1`)
	// The catalog, not an extra enum or hard-coded list, defines valid class roles.
	rows, err := s.pool.Query(ctx, `SELECT c.id,c.faction_id,r.role FROM classes c JOIN class_roles r ON r.class_id=c.id ORDER BY c.id,r.role`)
	if err != nil {
		t.Fatal(err)
	}
	type pair struct {
		classID, factionID int
		role               string
	}
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.classID, &p.factionID, &p.role); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		pairs = append(pairs, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 20 {
		t.Fatalf("expected seeded class roles, got %d", len(pairs))
	}
	for _, p := range pairs {
		mustExec(t, ctx, s, `UPDATE characters SET class_id=$1,faction_id=$2,role=$3 WHERE id=1`, p.classID, p.factionID, p.role)
	}
}

func TestLevelCatalogRejectsInvalidBounds(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, tc := range []struct{ sql, constraint string }{
		{`INSERT INTO levels(level,experience_required) VALUES(0,0)`, "levels_level_positive_check"},
		{`INSERT INTO levels(level,experience_required) VALUES(-1,0)`, "levels_level_positive_check"},
		{`UPDATE levels SET level=0 WHERE level=50`, "levels_level_positive_check"},
		{`INSERT INTO levels(level,experience_required) VALUES(51,-1)`, "levels_experience_nonnegative_check"},
		{`UPDATE levels SET experience_required=-1 WHERE level=50`, "levels_experience_nonnegative_check"},
	} {
		_, err := s.pool.Exec(ctx, tc.sql)
		requireConstraint(t, err, "23514", tc.constraint)
	}
}

func TestAllBaseClassesSupportNewCharacterDefaults(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 1, "all-class-defaults")
	catalog, err := r.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Classes) != 8 {
		t.Fatal("base class catalog changed")
	}
	for i, class := range catalog.Classes {
		ch, err := r.Create(ctx, actor, character.CreateInput{Name: "Starter " + string(rune('A'+i)), ClassID: class.ID, FactionID: class.FactionID})
		if err != nil {
			t.Fatalf("default creation for %s failed: %v", class.Code, err)
		}
		if ch.Role != "dd" || ch.Level != 1 || ch.Experience != 0 || ch.HP != 100 || ch.HPCurrent != 100 || ch.MP != 100 || ch.MPCurrent != 100 {
			t.Fatalf("default state changed for %s: %+v", class.Code, ch)
		}
	}
}

func TestCharacterCreationMissingDefaultCatalogRollsBack(t *testing.T) {
	for _, kind := range []string{"role", "level"} {
		t.Run(kind, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			r := db.NewCharacterRepository(s.pool)
			actor := repositoryActor(t, ctx, s, 1, "missing-default")
			var in character.CreateInput
			in.Name = "Missing catalog"
			if err := s.pool.QueryRow(ctx, `SELECT id,faction_id FROM classes WHERE code='bounty_hunter'`).Scan(&in.ClassID, &in.FactionID); err != nil {
				t.Fatal(err)
			}
			constraint := "characters_class_role_fkey"
			if kind == "role" {
				mustExec(t, ctx, s, `DELETE FROM class_roles WHERE class_id=$1 AND role='dd'`, in.ClassID)
			} else {
				mustExec(t, ctx, s, `UPDATE characters SET level=2; DELETE FROM levels WHERE level=1`)
				constraint = "characters_level_fkey"
			}
			mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
			before := characterStatsSnapshot(t, ctx, s)
			_, err := r.Create(ctx, actor, in)
			requireConstraint(t, err, "23503", constraint)
			if errors.Is(err, character.ErrInvalidClass) {
				t.Fatal("broken server catalog was reported as an invalid player class choice")
			}
			if characterStatsSnapshot(t, ctx, s) != before {
				t.Fatal("failed creation changed the account or left a partial character")
			}
		})
	}
}
