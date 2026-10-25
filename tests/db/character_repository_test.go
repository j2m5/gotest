package db_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/db"
)

func repositoryActor(t *testing.T, ctx context.Context, s *testStorage, id int, token string) character.Actor {
	t.Helper()
	if err := s.CreateSession(ctx, id, token, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return character.Actor{UserID: id, SessionHash: token}
}
func repositoryInput(t *testing.T, ctx context.Context, s *testStorage, name string) character.CreateInput {
	t.Helper()
	in := character.CreateInput{Name: name}
	if err := s.pool.QueryRow(ctx, `SELECT faction_id,id FROM classes WHERE code='sith_warrior'`).Scan(&in.FactionID, &in.ClassID); err != nil {
		t.Fatal(err)
	}
	return in
}
func TestCharacterRepositoryCatalogRosterAndCreation(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	a := repositoryActor(t, ctx, s, 1, "one")
	mustExec(t, ctx, s, `UPDATE factions SET description=NULL,icon=NULL; UPDATE classes SET description=NULL,icon=NULL`)
	catalog, err := r.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Factions) != 2 || len(catalog.Classes) != 8 {
		t.Fatalf("catalog=%+v", catalog)
	}
	for _, f := range catalog.Factions {
		if f.Description != nil || f.Icon != nil {
			t.Fatalf("nullable faction=%+v", f)
		}
	}
	for _, c := range catalog.Classes {
		if c.Description != nil || c.Icon != nil {
			t.Fatalf("nullable class=%+v", c)
		}
	}
	roster, err := r.List(ctx, a)
	if err != nil || len(roster.Characters) != 2 || roster.ActiveCharacterID != nil {
		t.Fatalf("roster=%+v err=%v", roster, err)
	}
	ch, err := r.Create(ctx, a, repositoryInput(t, ctx, s, "New warrior"))
	if err != nil {
		t.Fatal(err)
	}
	if ch.Role != "dd" || ch.Level != 1 || ch.Experience != 0 || ch.Avatar != nil || ch.HP != 100 || ch.MP != 100 || ch.HPCurrent != 100 || ch.MPCurrent != 100 {
		t.Fatalf("defaults=%+v", ch)
	}
	var got, want int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM character_abilities WHERE character_id=$1),(SELECT count(*) FROM class_abilities WHERE class_id=$2)`, ch.ID, ch.ClassID).Scan(&got, &want); err != nil {
		t.Fatal(err)
	}
	if want == 0 || got != want {
		t.Fatalf("abilities=%d want=%d", got, want)
	}
	var mismatch bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS((SELECT ability_id FROM character_abilities WHERE character_id=$1 EXCEPT SELECT ability_id FROM class_abilities WHERE class_id=$2) UNION ALL (SELECT ability_id FROM class_abilities WHERE class_id=$2 EXCEPT SELECT ability_id FROM character_abilities WHERE character_id=$1))`, ch.ID, ch.ClassID).Scan(&mismatch); err != nil || mismatch {
		t.Fatalf("ability membership mismatch=%v error=%v", mismatch, err)
	}
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("creation selected character")
	}
	selected, err := r.SetActive(ctx, a, &ch.ID)
	if err != nil || selected.ID != ch.ID {
		t.Fatalf("selection=%+v %v", selected, err)
	}
	roster, err = r.List(ctx, a)
	if err != nil || roster.ActiveCharacterID == nil || *roster.ActiveCharacterID != ch.ID {
		t.Fatalf("selected roster=%+v %v", roster, err)
	}
	for _, id := range []int{3, 99999} {
		if _, err := r.SetActive(ctx, a, &id); !errors.Is(err, character.ErrNotFound) {
			t.Fatalf("foreign/missing=%v", err)
		}
	}
	if ch, err := r.SetActive(ctx, a, nil); err != nil || ch != nil {
		t.Fatalf("clear=%+v %v", ch, err)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=NULL; DELETE FROM characters WHERE user_id=1`)
	roster, err = r.List(ctx, a)
	if err != nil || roster.Characters == nil || len(roster.Characters) != 0 {
		t.Fatalf("empty=%+v %v", roster, err)
	}
}
func TestCharacterRepositoryRejectsInvalidActorsAndClasses(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	valid := repositoryActor(t, ctx, s, 1, "live")
	expired := repositoryActor(t, ctx, s, 1, "expired")
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token='expired'`)
	revoked := repositoryActor(t, ctx, s, 1, "revoked")
	mustExec(t, ctx, s, `DELETE FROM sessions WHERE token='revoked'`)
	unverified := repositoryActor(t, ctx, s, 2, "unverified")
	mustExec(t, ctx, s, `UPDATE users SET email_verified_at=NULL WHERE id=2`)
	for _, a := range []character.Actor{{}, expired, revoked, unverified, {UserID: 2, SessionHash: valid.SessionHash}} {
		if _, err := r.List(ctx, a); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("list actor=%+v err=%v", a, err)
		}
		if _, err := r.Create(ctx, a, repositoryInput(t, ctx, s, "Forbidden")); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("create err=%v", err)
		}
		if _, err := r.SetActive(ctx, a, nil); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("clear err=%v", err)
		}
	}
	for _, in := range []character.CreateInput{{Name: "Bad pair", FactionID: 999, ClassID: 1}, {Name: "Bad class", FactionID: 1, ClassID: 999}} {
		if _, err := r.Create(ctx, valid, in); !errors.Is(err, character.ErrInvalidClass) {
			t.Fatalf("invalid class=%v", err)
		}
	}
	in := repositoryInput(t, ctx, s, "Bad faction")
	in.FactionID = 3 - in.FactionID
	if _, err := r.Create(ctx, valid, in); !errors.Is(err, character.ErrInvalidClass) {
		t.Fatalf("mismatched class=%v", err)
	}
}
func TestCharacterRepositoryAbilityFailureRollsBack(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	a := repositoryActor(t, ctx, s, 1, "live")
	mustExec(t, ctx, s, `CREATE FUNCTION reject_ability() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected ability failure'; END $$; CREATE TRIGGER reject_ability BEFORE INSERT ON character_abilities FOR EACH ROW EXECUTE FUNCTION reject_ability()`)
	if _, err := r.Create(ctx, a, repositoryInput(t, ctx, s, "Rollback warrior")); err == nil {
		t.Fatal("expected ability failure")
	}
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM characters WHERE name='Rollback warrior'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("half-created count=%d err=%v", n, err)
	}
}
func TestCharacterRepositoryConcurrentNormalizedNames(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	actors := []character.Actor{repositoryActor(t, ctx, s, 1, "one"), repositoryActor(t, ctx, s, 2, "two")}
	inputs := []character.CreateInput{repositoryInput(t, ctx, s, "Caf\u00e9 Warrior"), repositoryInput(t, ctx, s, "  CAFE\u0301  WARRIOR  ")}
	results := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; _, err := r.Create(ctx, actors[i], inputs[i]); results <- err }()
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, character.ErrNameTaken) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}
func TestCharacterNamesMigrationRejectsDuplicatesAndReverses(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	executeMigration(t, ctx, s, "20261025000400_character_names.down.sql")
	mustExec(t, ctx, s, `INSERT INTO characters(user_id,faction_id,class_id,name) SELECT 1,faction_id,id,'  FIRST  CHARACTER ' FROM classes LIMIT 1`)
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, migrationSQL(t, "20261025000400_character_names.up.sql")); err == nil {
		t.Fatal("duplicate migration succeeded")
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM characters`).Scan(&n); err != nil || n != 4 {
		t.Fatalf("data lost count=%d err=%v", n, err)
	}
	if _, err := conn.Exec(ctx, `DELETE FROM characters WHERE name='  FIRST  CHARACTER '`); err != nil {
		t.Fatal(err)
	}
	executeMigration(t, ctx, s, "20261025000400_character_names.up.sql")
	executeMigration(t, ctx, s, "20261025000400_character_names.down.sql")
	mustExec(t, ctx, s, `INSERT INTO characters(user_id,faction_id,class_id,name) SELECT 1,faction_id,id,'FIRST CHARACTER' FROM classes LIMIT 1`)
}
func TestCharacterRepositoryRechecksRevocationAfterAccountLock(t *testing.T) {
	for _, operation := range []string{"create", "select", "clear"} {
		t.Run(operation, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			r := db.NewCharacterRepository(s.pool)
			a := repositoryActor(t, ctx, s, 1, "live")
			in := repositoryInput(t, ctx, s, "Stale creation")
			blocker, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Rollback(ctx)
			var pid int
			if err := blocker.QueryRow(ctx, `SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE`).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				var err error
				switch operation {
				case "create":
					_, err = r.Create(ctx, a, in)
				case "select":
					id := 1
					_, err = r.SetActive(ctx, a, &id)
				default:
					_, err = r.SetActive(ctx, a, nil)
				}
				done <- err
			}()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				var waiting bool
				if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
					t.Fatal(err)
				}
				if waiting {
					break
				}
				select {
				case err := <-done:
					t.Fatalf("did not wait for account lock: %v", err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-ticker.C:
				}
			}
			if _, err := blocker.Exec(ctx, `DELETE FROM sessions WHERE token=$1`, a.SessionHash); err != nil {
				t.Fatal(err)
			}
			if _, err := blocker.Exec(ctx, `UPDATE users SET character_id=2 WHERE id=1`); err != nil {
				t.Fatal(err)
			}
			if err := blocker.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, character.ErrUnauthenticated) {
				t.Fatalf("stale mutation=%v", err)
			}
			if got := activeCharacter(t, ctx, s, 1); got == nil || *got != 2 {
				t.Fatal("revoked actor changed selection")
			}
			var n int
			if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM characters WHERE name='Stale creation'`).Scan(&n); err != nil || n != 0 {
				t.Fatalf("stale creation count=%d err=%v", n, err)
			}
		})
	}
}

func TestCharacterNameKeyMatchesUnicodeNormalization(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, pair := range [][2]string{
		{" \tCaf\u00e9  Warrior\u3000", "CAFE\u0301 WARRIOR"},
		{"\u0414\u0430\u0440\u0442  \u0412\u0435\u0439\u0434\u0435\u0440", "\u0434\u0430\u0440\u0442 \u0432\u0435\u0439\u0434\u0435\u0440"},
		{"\u00a0\u0085Hero\u202f", "hero"},
	} {
		var equal bool
		if err := s.pool.QueryRow(ctx, `SELECT character_name_key($1)=character_name_key($2)`, pair[0], pair[1]).Scan(&equal); err != nil {
			t.Fatal(err)
		}
		if !equal {
			t.Fatalf("keys differ: %q and %q", pair[0], pair[1])
		}
	}
}
func TestCharacterRepositoryChecksCurrentExpiryAfterAccountLock(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	a := repositoryActor(t, ctx, s, 1, "expiring")
	blocker, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback(ctx)
	var pid int
	if err := blocker.QueryRow(ctx, `SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { id := 1; _, err := r.SetActive(ctx, a, &id); done <- err }()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("did not wait: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
	// This timestamp is later than the waiting transaction's now(), but expires
	// before it acquires the user row. Only wall-clock expiry rejects the request.
	if _, err := blocker.Exec(ctx, `UPDATE sessions SET expires_at=clock_timestamp() WHERE token=$1`, a.SessionHash); err != nil {
		t.Fatal(err)
	}
	if err := blocker.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("expired waiter=%v", err)
	}
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("expired waiter selected a character")
	}
}

func TestCharacterNameKeyIgnoresCallerCollation(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, pair := range [][2]string{
		{"\u0414\u0410\u0420\u0422", "\u0434\u0430\u0440\u0442"},
		{"CAF\u00c9", "caf\u00e9"},
	} {
		var equal bool
		if err := s.pool.QueryRow(ctx, `SELECT character_name_key($1 COLLATE "C")=character_name_key($2 COLLATE "C")`, pair[0], pair[1]).Scan(&equal); err != nil {
			t.Fatal(err)
		}
		if !equal {
			t.Fatalf("caller C collation bypasses Unicode case uniqueness for %q and %q", pair[0], pair[1])
		}
	}
}

func TestCharacterNameKeyPinsUnicodeCollation(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	// The parsed SQL body must bind Unicode case conversion independently of
	// database LC_CTYPE. Checking its dependency also covers C-default databases
	// without creating or changing a database in this integration fixture.
	var pinned bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM pg_depend d JOIN pg_collation c ON c.oid=d.refobjid
 WHERE d.classid='pg_proc'::regclass AND d.objid='character_name_key(text)'::regprocedure
 AND d.refclassid='pg_collation'::regclass AND c.collname='character_name_unicode'
 AND c.collnamespace=current_schema()::regnamespace AND c.collprovider='i')`).Scan(&pinned); err != nil {
		t.Fatal(err)
	}
	if !pinned {
		t.Fatal("name key must pin its own Unicode ICU collation")
	}
}
