package db_test

import (
	"context"
	"errors"
	"testing"

	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/world"
)

func travelFixture(t *testing.T) (*testStorage, context.Context, *world.Service, character.Actor, int) {
	t.Helper()
	s, ctx := characterTestStorage(t, true)
	a := repositoryActor(t, ctx, s, 1, "travel-session")
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
	var location int
	if err := s.pool.QueryRow(ctx, `SELECT location_id FROM characters WHERE id=1`).Scan(&location); err != nil {
		t.Fatal(err)
	}
	return s, ctx, world.NewService(db.NewWorldRepository(s.pool)), a, location
}
func planetID(t *testing.T, ctx context.Context, s *testStorage, code string) int {
	t.Helper()
	var id int
	if err := s.pool.QueryRow(ctx, `SELECT id FROM planets WHERE code=$1`, code).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// Empty-world scenarios must not depend on which catalog planets have been filled.
func emptyCorePlanet(t *testing.T, ctx context.Context, s *testStorage) int {
	t.Helper()
	var id int
	if err := s.pool.QueryRow(ctx, `INSERT INTO planets(code,name,name_en,core_faction_id,controlling_faction_id)
		SELECT 'test_empty_core','Test empty core','Test empty core',id,id FROM factions WHERE code='sith_empire'
		RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func emptyNeutralPlanet(t *testing.T, ctx context.Context, s *testStorage) int {
	t.Helper()
	var id int
	if err := s.pool.QueryRow(ctx, `INSERT INTO planets(code,name,name_en)
		VALUES('test_empty_neutral','Test empty neutral','Test empty neutral') RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestWorldNavigationAndTravelPersistLowestLocation(t *testing.T) {
	s, ctx, service, actor, start := travelFixture(t)
	nav, err := service.Navigation(ctx, actor, 1)
	if err != nil || nav.Position.Planet.Code != "dromund_kaas" || nav.Position.Location.ID != start || len(nav.Destinations) != 100 {
		t.Fatalf("navigation: %+v %v", nav, err)
	}
	planet := emptyCorePlanet(t, ctx, s)
	mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'higher','Higher first',1,5),($1,'low','Lowest first',2,1),($1,'tie','Lowest second',3,1)`, planet)
	input := world.TravelInput{CharacterID: 1, PlanetID: planet, FromLocationID: start}
	result, err := service.Travel(ctx, actor, input)
	if err != nil || result.Position.Location.Code != "low" || result.ReturnedHome {
		t.Fatalf("arrival: %+v %v", result, err)
	}
	_, err = db.NewCharacterRepository(s.pool).SetActive(ctx, actor, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.NewCharacterRepository(s.pool).SetActive(ctx, actor, new(1))
	if err != nil {
		t.Fatal(err)
	}
	nav, err = service.Navigation(ctx, actor, 1)
	if err != nil || nav.Position.Location.ID != result.Position.Location.ID {
		t.Fatalf("selection reset position: %+v %v", nav, err)
	}
	// A duplicate flight to the current planet must not overwrite local movement.
	mustExec(t, ctx, s, `UPDATE characters SET location_id=(SELECT id FROM locations WHERE planet_id=$1 AND code='tie') WHERE id=1`, planet)
	result, err = service.Travel(ctx, actor, input)
	if err != nil || result.Position.Location.Code != "tie" {
		t.Fatalf("duplicate flight relocated character: %+v %v", result, err)
	}
	_, err = service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: planetID(t, ctx, s, "dromund_kaas"), FromLocationID: start})
	if !errors.Is(err, world.ErrPositionChanged) {
		t.Fatalf("stale position accepted: %v", err)
	}
}

func TestWorldTravelAccessAndLostControl(t *testing.T) {
	s, ctx, service, actor, start := travelFixture(t)
	emptyCorePlanet(t, ctx, s)
	for code, want := range map[string]error{"coruscant": world.ErrAccessDenied, "tatooine": world.ErrAccessDenied, "test_empty_core": world.ErrLocationsUnavailable} {
		_, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: planetID(t, ctx, s, code), FromLocationID: start})
		if !errors.Is(err, want) {
			t.Fatalf("%s: %v", code, err)
		}
	}
	id := emptyNeutralPlanet(t, ctx, s)
	mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'test_area','Test area',1,1);`, id)
	mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT faction_id FROM characters WHERE id=1) WHERE id=$1`, id)
	moved, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: id, FromLocationID: start})
	if err != nil || moved.Position.Planet.ID != id {
		t.Fatalf("owned noncore unavailable: %+v %v", moved, err)
	}
	mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=NULL WHERE id=$1`, id)
	nav, err := service.Navigation(ctx, actor, 1)
	if err != nil || !nav.ReturnRequired || nav.Position.Planet.ID != id {
		t.Fatalf("GET mutated lost-world position: %+v %v", nav, err)
	}
	if _, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: id, FromLocationID: start}); !errors.Is(err, world.ErrPositionChanged) {
		t.Fatalf("stale action returned from lost world: %v", err)
	}
	var actual int
	if err := s.pool.QueryRow(ctx, `SELECT location_id FROM characters WHERE id=1`).Scan(&actual); err != nil || actual != moved.Position.Location.ID {
		t.Fatalf("stale action moved character: %d %v", actual, err)
	}
	returned, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: id, FromLocationID: moved.Position.Location.ID})
	if err != nil || !returned.ReturnedHome || returned.Position.Planet.Code != "dromund_kaas" {
		t.Fatalf("lost world return: %+v %v", returned, err)
	}
}

func TestWorldTravelQueuedRequestRechecksAccount(t *testing.T) {
	for _, change := range []string{"session", "selection", "ban"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, service, actor, start := travelFixture(t)
			input := world.TravelInput{CharacterID: 1, PlanetID: planetID(t, ctx, s, "korriban"), FromLocationID: start}
			tx, pid := lockResetUser(t, ctx, s)
			done := make(chan error, 1)
			go func() { _, err := service.Travel(ctx, actor, input); done <- err }()
			awaitResetUserLock(t, ctx, s, pid, done)
			sql := `DELETE FROM sessions WHERE user_id=1`
			want := character.ErrUnauthenticated
			if change == "selection" {
				sql = `UPDATE users SET character_id=2 WHERE id=1`
				want = world.ErrCharacterChanged
			}
			if change == "ban" {
				sql = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`
			}
			if _, err := tx.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, want) {
				t.Fatalf("stale %s: %v", change, err)
			}
			var actual int
			if err := s.pool.QueryRow(ctx, `SELECT location_id FROM characters WHERE id=1`).Scan(&actual); err != nil || actual != start {
				t.Fatalf("unauthorized move persisted: %d %v", actual, err)
			}
		})
	}
}

func TestWorldTravelRechecksAfterPlanetLock(t *testing.T) {
	for _, change := range []string{"control", "expiry"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, service, actor, start := travelFixture(t)
			id := emptyNeutralPlanet(t, ctx, s)
			mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'test_area','Test area',1,1)`, id)
			mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT faction_id FROM characters WHERE id=1) WHERE id=$1`, id)
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			var pid int
			if err := tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM planets WHERE id=$1 FOR UPDATE`, id).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				_, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: id, FromLocationID: start})
				done <- err
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			if change == "control" {
				_, err = tx.Exec(ctx, `UPDATE planets SET controlling_faction_id=NULL WHERE id=$1`, id)
			} else {
				_, err = tx.Exec(ctx, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, actor.SessionHash)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			want := world.ErrAccessDenied
			if change == "expiry" {
				want = character.ErrUnauthenticated
			}
			if err = <-done; !errors.Is(err, want) {
				t.Fatalf("stale %s after lock: %v", change, err)
			}
		})
	}
}
