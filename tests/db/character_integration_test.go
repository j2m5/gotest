package db_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gotest/internal/auth"
	"gotest/internal/models"
	"gotest/internal/seed"
)

const characterMigration = "20261025000300_character_integrity"

func characterTestStorage(t *testing.T, latest bool) (*testStorage, context.Context) {
	t.Helper()
	s, ctx := authTestStorage(t, true)
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		name := filepath.Base(path)
		switch name {
		case "000001_create_users.up.sql", "20260514075823_create_sessions.up.sql", "20260516144648_create_email_verifications.up.sql", "20261025000100_auth_hardening.up.sql":
			continue // Applied by the auth fixture.
		}
		if !latest && name >= characterMigration+".up.sql" {
			continue
		}
		executeMigration(t, ctx, s, name)
	}
	if _, err := seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at) VALUES
		('first@example.test','first','unused',NOW()), ('second@example.test','second','unused',NOW());
		INSERT INTO characters(user_id,faction_id,class_id,name)
		SELECT 1,faction_id,id,'First character' FROM classes WHERE code='sith_warrior';
		INSERT INTO characters(user_id,faction_id,class_id,name)
		SELECT 1,faction_id,id,'Second character' FROM classes WHERE code='jedi_consular';
		INSERT INTO characters(user_id,faction_id,class_id,name)
		SELECT 2,faction_id,id,'Other account character' FROM classes WHERE code='trooper'`)
	return s, ctx
}

func mustExec(t *testing.T, ctx context.Context, s *testStorage, sql string, args ...any) {
	t.Helper()
	if _, err := s.pool.Exec(ctx, sql, args...); err != nil {
		t.Fatal(err)
	}
}

func wantForeignKey(t *testing.T, err error) {
	t.Helper()
	var postgres *pgconn.PgError
	if !errors.As(err, &postgres) || postgres.Code != "23503" {
		t.Fatalf("expected foreign-key rejection, got %v", err)
	}
}

func activeCharacter(t *testing.T, ctx context.Context, s *testStorage, userID int) *int {
	t.Helper()
	var id *int
	if err := s.pool.QueryRow(ctx, "SELECT character_id FROM users WHERE id=$1", userID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCharacterClassFactionIntegrity(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, sql := range []string{
		`INSERT INTO characters(user_id,faction_id,class_id,name) SELECT 1,f.id,c.id,'Invalid pair' FROM factions f,classes c WHERE f.code='galactic_republic' AND c.code='sith_warrior'`,
		`UPDATE characters SET faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE id=1`,
		`UPDATE characters SET class_id=(SELECT id FROM classes WHERE code='trooper') WHERE id=1`,
		`UPDATE classes SET faction_id=(SELECT id FROM factions WHERE code='galactic_republic') WHERE code='sith_warrior'`,
	} {
		_, err := s.pool.Exec(ctx, sql)
		wantForeignKey(t, err)
	}
	// Both fields may change together to a valid pair.
	mustExec(t, ctx, s, `UPDATE characters SET (class_id,faction_id)=(SELECT id,faction_id FROM classes WHERE code='jedi_knight') WHERE id=1`)
}

func TestCharacterCatalogHasFourBaseClassesPerFaction(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for faction, want := range map[string][]string{
		"sith_empire":       {"bounty_hunter", "imperial_agent", "sith_inquisitor", "sith_warrior"},
		"galactic_republic": {"jedi_consular", "jedi_knight", "smuggler", "trooper"},
	} {
		var got []string
		if err := s.pool.QueryRow(ctx, `SELECT array_agg(c.code ORDER BY c.code) FROM classes c JOIN factions f ON f.id=c.faction_id WHERE f.code=$1`, faction).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("classes of %s: got %v, want %v", faction, got, want)
		}
	}
}

func TestActiveCharacterMustBelongToAccount(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("new account must have no active character")
	}
	for _, id := range []int{3, 999999} {
		_, err := s.pool.Exec(ctx, "UPDATE users SET character_id=$1 WHERE id=1", id)
		wantForeignKey(t, err)
	}
	for _, id := range []int{1, 2} {
		mustExec(t, ctx, s, "UPDATE users SET character_id=$1 WHERE id=1", id)
		if got := activeCharacter(t, ctx, s, 1); got == nil || *got != id {
			t.Fatalf("own character not selected: %v", got)
		}
	}
	mustExec(t, ctx, s, "UPDATE users SET character_id=NULL WHERE id=1")
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("cannot return to character selection")
	}
}

func TestActiveCharacterDeletionAndAccountCascade(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, "UPDATE users SET character_id=1 WHERE id=1")
	_, err := s.pool.Exec(ctx, "DELETE FROM characters WHERE id=1")
	wantForeignKey(t, err)
	_, err = s.pool.Exec(ctx, "UPDATE characters SET user_id=2 WHERE id=1")
	wantForeignKey(t, err)
	mustExec(t, ctx, s, "UPDATE users SET character_id=NULL WHERE id=1; DELETE FROM characters WHERE id=1")
	mustExec(t, ctx, s, "UPDATE users SET character_id=2 WHERE id=1; DELETE FROM users WHERE id=1")
	var own, other int
	if err := s.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM characters WHERE user_id=1),(SELECT count(*) FROM characters WHERE user_id=2)").Scan(&own, &other); err != nil || own != 0 || other != 1 {
		t.Fatalf("account cascade: own=%d other=%d err=%v", own, other, err)
	}
}

func TestUserQueriesLoadNullableActiveCharacter(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	if err := s.CreateSession(ctx, 1, "lookup", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []*int{nil, new(2)} {
		mustExec(t, ctx, s, "UPDATE users SET character_id=$1 WHERE id=1", want)
		u, err := s.FindUserByEmail(ctx, "first@example.test")
		if err != nil || !reflect.DeepEqual(u.CharacterID, want) {
			t.Fatalf("FindUserByEmail active character: %+v, want %v, err=%v", u, want, err)
		}
		session, err := s.FindSession(ctx, "lookup")
		if err != nil || !reflect.DeepEqual(session.User.CharacterID, want) {
			t.Fatalf("FindSession active character: %+v, want %v, err=%v", session, want, err)
		}
	}
}

func TestNullableCatalogAndCharacterModels(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	t.Run("faction", func(t *testing.T) {
		mustExec(t, ctx, s, "UPDATE factions SET description=NULL,icon=NULL WHERE code='sith_empire'")
		var f models.Faction
		if err := s.pool.QueryRow(ctx, "SELECT description,icon FROM factions WHERE code='sith_empire'").Scan(&f.Description, &f.Icon); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(f.Description, (*string)(nil)) || !reflect.DeepEqual(f.Icon, (*string)(nil)) {
			t.Fatal("NULL fields must remain absent")
		}
	})
	t.Run("class", func(t *testing.T) {
		mustExec(t, ctx, s, "UPDATE classes SET alias=NULL,description=NULL,icon=NULL WHERE code='sith_warrior'")
		var c models.Class
		if err := s.pool.QueryRow(ctx, "SELECT alias,description,icon FROM classes WHERE code='sith_warrior'").Scan(&c.Alias, &c.Description, &c.Icon); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Alias, (*string)(nil)) || !reflect.DeepEqual(c.Description, (*string)(nil)) || !reflect.DeepEqual(c.Icon, (*string)(nil)) {
			t.Fatal("NULL fields must remain absent")
		}
		mustExec(t, ctx, s, "UPDATE classes SET alias='Warrior',description='',icon='warrior.png' WHERE code='sith_warrior'")
		if err := s.pool.QueryRow(ctx, "SELECT alias,description,icon FROM classes WHERE code='sith_warrior'").Scan(&c.Alias, &c.Description, &c.Icon); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Alias, new("Warrior")) || !reflect.DeepEqual(c.Description, new("")) || !reflect.DeepEqual(c.Icon, new("warrior.png")) {
			t.Fatal("present strings, including empty string, must remain distinct from NULL")
		}
	})
	t.Run("ability", func(t *testing.T) {
		mustExec(t, ctx, s, "UPDATE abilities SET description=NULL,icon=NULL WHERE code='sith_warrior_basic_attack'")
		var a models.Ability
		if err := s.pool.QueryRow(ctx, "SELECT description,icon FROM abilities WHERE code='sith_warrior_basic_attack'").Scan(&a.Description, &a.Icon); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a.Description, (*string)(nil)) || !reflect.DeepEqual(a.Icon, (*string)(nil)) {
			t.Fatal("NULL fields must remain absent")
		}
	})
	t.Run("character", func(t *testing.T) {
		var c models.Character
		if err := s.pool.QueryRow(ctx, "SELECT avatar,activated_at FROM characters WHERE id=1").Scan(&c.Avatar, &c.ActivatedAt); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Avatar, (*string)(nil)) || c.ActivatedAt != nil {
			t.Fatal("NULL fields must remain absent")
		}
	})
}

func sessionFor(t *testing.T, ctx context.Context, s *testStorage, userID int, marker string, expiry time.Time) string {
	t.Helper()
	raw := strings.Repeat(marker, 64)
	hash := sha256.Sum256([]byte(raw))
	if err := s.CreateSession(ctx, userID, hex.EncodeToString(hash[:]), expiry); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestLogoutClearsSharedSelectionAndPreservesOtherSessions(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s, nil, auth.Options{})
	first := sessionFor(t, ctx, s, 1, "a", time.Now().Add(time.Hour))
	otherDevice := sessionFor(t, ctx, s, 1, "b", time.Now().Add(time.Hour))
	otherAccount := sessionFor(t, ctx, s, 2, "c", time.Now().Add(time.Hour))
	mustExec(t, ctx, s, "UPDATE users SET character_id=1 WHERE id=1; UPDATE users SET character_id=3 WHERE id=2")
	if err := service.Logout(ctx, first); err != nil {
		t.Fatal(err)
	}
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("logout retained the account's active character")
	}
	if _, err := service.CurrentUser(ctx, first); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("logged out session still authenticates: %v", err)
	}
	if _, err := service.CurrentUser(ctx, otherDevice); err != nil {
		t.Fatalf("logout revoked another device: %v", err)
	}
	if _, err := service.CurrentUser(ctx, otherAccount); err != nil {
		t.Fatal(err)
	}
	if got := activeCharacter(t, ctx, s, 2); got == nil || *got != 3 {
		t.Fatal("logout changed another account's selection")
	}
	mustExec(t, ctx, s, "UPDATE users SET character_id=2 WHERE id=1")
	for _, old := range []string{first, "", strings.Repeat("d", 64)} {
		if err := service.Logout(ctx, old); err != nil {
			t.Fatal(err)
		}
	}
	if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 2 {
		t.Fatal("replaying logout cleared a later character selection")
	}
}

func TestExpiredLogoutCannotChangeSelection(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s, nil, auth.Options{})
	raw := sessionFor(t, ctx, s, 1, "a", time.Now().Add(-time.Hour))
	mustExec(t, ctx, s, "UPDATE users SET character_id=1 WHERE id=1")
	if err := service.Logout(ctx, raw); err != nil {
		t.Fatal(err)
	}
	if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 1 {
		t.Fatal("expired credentials changed the active character")
	}
	var sessions int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id=1").Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("expired session was not removed: %d, %v", sessions, err)
	}
}

func TestLogoutRollsBackRevocationIfSelectionResetFails(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s, nil, auth.Options{})
	raw := sessionFor(t, ctx, s, 1, "a", time.Now().Add(time.Hour))
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1;
		CREATE FUNCTION reject_selection_reset() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected reset failure'; END $$;
		CREATE TRIGGER reject_selection_reset BEFORE UPDATE OF character_id ON users FOR EACH ROW EXECUTE FUNCTION reject_selection_reset()`)
	if err := service.Logout(ctx, raw); err == nil {
		t.Fatal("logout must fail if resetting selection fails")
	}
	if _, err := service.CurrentUser(ctx, raw); err != nil {
		t.Fatalf("failed logout lost its session: %v", err)
	}
	if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 1 {
		t.Fatal("failed logout changed selection")
	}
}

func TestLogoutRechecksSessionAfterWaitingForAccountLock(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := auth.NewService(s, nil, auth.Options{})
	raw := sessionFor(t, ctx, s, 1, "a", time.Now().Add(time.Hour))
	mustExec(t, ctx, s, "UPDATE users SET character_id=1 WHERE id=1")
	blocker, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback(ctx)
	var blockerPID int
	if err := blocker.QueryRow(ctx, "SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE").Scan(&blockerPID); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- service.Logout(ctx, raw) }()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("logout did not serialize with the account lock: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
	// Another logout consumed this token, then a live device selected character 2.
	hash := sha256.Sum256([]byte(raw))
	if _, err := blocker.Exec(ctx, "DELETE FROM sessions WHERE token=$1", hex.EncodeToString(hash[:])); err != nil {
		t.Fatal(err)
	}
	if _, err := blocker.Exec(ctx, "UPDATE users SET character_id=2 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if err := blocker.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 2 {
		t.Fatal("stale logout cleared a selection made after its session was revoked")
	}
}
