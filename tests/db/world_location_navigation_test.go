package db_test

import (
	"context"
	"errors"
	"gotest/internal/character"
	"gotest/internal/seed"
	"gotest/internal/world"
	"path/filepath"
	"testing"
)

func TestPopulatedWorldLocationContent(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, code := range []string{
		"coruscant", "dromund_kaas", "tython", "korriban",
		"arkania", "balmorra", "ord_mantell", "athiss", "thule", "ossus",
		"carida", "cato_neimoidia", "shili", "ord_radama", "mirial", "telos_iv",
		"commenor", "ralltiir", "humbarine", "neimoidia", "dathomir", "saleucami", "sriluur", "dac",
		"corellia", "alderaan", "kuat", "ziost", "begeren", "bosthirda",
		"chandrila", "duro", "brentaal_iv", "jaguada", "rhelg", "khar_delba",
		"anaxes", "alsakan", "rendili", "dromund_fels", "krayiss_ii", "chhodos",
		"caamas", "esseles", "rhinnal", "corulag", "ashas_ree", "kalsunor", "korriz", "upekzar",
		"belsavis", "bothawui", "dantooine", "taris", "hoth", "tatooine", "manaan", "onderon",
		"makeb", "nal_hutta", "kashyyyk", "trandosha", "ryloth", "rodia", "sullust", "ithor",
		"ilum", "rishi", "mandalore", "concord_dawn", "cathar", "myrkr", "iridonia", "ambria",
		"sleheyron", "toydaria", "kwenn", "nar_kreeta", "klatooine", "vodran", "kintan", "ubrikkia",
		"dorin", "glee_anselm", "gand", "togoria", "eres_iii", "taanab", "zeltros", "cerea", "bimmisaari",
		"malastare", "kessel", "eriadu", "sluis_van", "fondor", "bestine_iv", "denon", "mygeeto", "muunilinst",
	} {
		var levels, positions []int
		var descriptions int
		err := s.pool.QueryRow(ctx, `SELECT array_agg(l.recommended_level ORDER BY l.position),array_agg(l.position ORDER BY l.position),count(*) FILTER(WHERE btrim(l.description)<>'') FROM locations l JOIN planets p ON p.id=l.planet_id WHERE p.code=$1`, code).Scan(&levels, &positions, &descriptions)
		if err != nil || len(levels) != 10 || descriptions != 10 {
			t.Fatalf("%s content: levels=%v descriptions=%d err=%v", code, levels, descriptions, err)
		}
		for i, level := range levels {
			if level != 1+5*i || positions[i] != i+1 {
				t.Fatalf("%s progression: %v %v", code, levels, positions)
			}
		}
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM locations`).Scan(&count); err != nil || count != 1000 {
		t.Fatalf("location count %d: %v", count, err)
	}
	mustExec(t, ctx, s, `UPDATE locations SET description='User-authored description',recommended_level=2 WHERE id=(SELECT location_id FROM characters WHERE id=1)`)
	var before, after string
	snapshot := `SELECT jsonb_build_object('locations',(SELECT jsonb_agg(to_jsonb(l) ORDER BY id) FROM locations l),'positions',(SELECT jsonb_agg(location_id ORDER BY id) FROM characters))::text`
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Run(ctx, s.pool, filepath.Join("..", "..", "seeds")); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || after != before {
		t.Fatalf("reseed changed content or positions: %v", err)
	}
}

func locationAt(t *testing.T, ctx context.Context, s *testStorage, planet string, position int) int {
	t.Helper()
	var id int
	if err := s.pool.QueryRow(ctx, `SELECT l.id FROM locations l JOIN planets p ON p.id=l.planet_id WHERE p.code=$1 AND l.position=$2`, planet, position).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestWorldLocationsAndMovement(t *testing.T) {
	s, ctx, service, actor, start := travelFixture(t)
	home := planetID(t, ctx, s, "dromund_kaas")
	page, err := service.Locations(ctx, actor, 1, home)
	if err != nil || len(page.Locations) != 10 || page.Position.Location.ID != start || page.ReturnRequired {
		t.Fatalf("locations: %+v %v", page, err)
	}
	target := page.Locations[9]
	if target.RecommendedLevel != 46 {
		t.Fatalf("last recommendation: %+v", target)
	}
	in := world.MoveInput{CharacterID: 1, LocationID: target.ID, FromLocationID: start}
	moved, err := service.Move(ctx, actor, in)
	if err != nil || moved.Position.Location.ID != target.ID || moved.ReturnedHome {
		t.Fatalf("move level1 into level46: %+v %v", moved, err)
	}
	if _, err = service.Move(ctx, actor, in); err != nil {
		t.Fatalf("duplicate move: %v", err)
	}
	var actual, level int
	if err = s.pool.QueryRow(ctx, `SELECT location_id,level FROM characters WHERE id=1`).Scan(&actual, &level); err != nil || actual != target.ID || level != 1 {
		t.Fatalf("position/level: %d/%d %v", actual, level, err)
	}
	page, err = service.Locations(ctx, actor, 1, home)
	if err != nil || page.Position.Location.ID != target.ID {
		t.Fatalf("read moved position: %+v %v", page, err)
	}
	cases := []struct {
		input world.MoveInput
		want  error
	}{
		{world.MoveInput{CharacterID: 1, LocationID: start, FromLocationID: start}, world.ErrPositionChanged},
		{world.MoveInput{CharacterID: 1, LocationID: locationAt(t, ctx, s, "korriban", 1), FromLocationID: target.ID}, world.ErrWrongPlanet},
		{world.MoveInput{CharacterID: 1, LocationID: 2147483647, FromLocationID: target.ID}, world.ErrLocationNotFound},
		{world.MoveInput{CharacterID: 2, LocationID: start, FromLocationID: target.ID}, world.ErrCharacterChanged},
	}
	for _, tc := range cases {
		if _, err = service.Move(ctx, actor, tc.input); !errors.Is(err, tc.want) {
			t.Fatalf("%+v: %v want%v", tc.input, err, tc.want)
		}
	}
	if _, err = service.Locations(ctx, actor, 1, planetID(t, ctx, s, "korriban")); !errors.Is(err, world.ErrWrongPlanet) {
		t.Fatalf("remote listing: %v", err)
	}
	if err = s.pool.QueryRow(ctx, `SELECT location_id FROM characters WHERE id=1`).Scan(&actual); err != nil || actual != target.ID {
		t.Fatalf("read/rejected action changed position: %d %v", actual, err)
	}
}

func TestWorldMoveQueuedAccountChanges(t *testing.T) {
	for _, change := range []string{"session", "selection", "ban"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, service, actor, start := travelFixture(t)
			target := locationAt(t, ctx, s, "dromund_kaas", 10)
			tx, pid := lockResetUser(t, ctx, s)
			done := make(chan error, 1)
			go func() {
				_, err := service.Move(ctx, actor, world.MoveInput{CharacterID: 1, LocationID: target, FromLocationID: start})
				done <- err
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			query := `DELETE FROM sessions WHERE user_id=1`
			want := character.ErrUnauthenticated
			if change == "selection" {
				query = `UPDATE users SET character_id=2 WHERE id=1`
				want = world.ErrCharacterChanged
			}
			if change == "ban" {
				query = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`
			}
			if _, err := tx.Exec(ctx, query); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, want) {
				t.Fatalf("queued %s: %v", change, err)
			}
			var actual int
			if err := s.pool.QueryRow(ctx, `SELECT location_id FROM characters WHERE id=1`).Scan(&actual); err != nil || actual != start {
				t.Fatalf("unauthorized move: %d %v", actual, err)
			}
		})
	}
}

func TestWorldLocalMoveAndFlightSerialize(t *testing.T) {
	s, ctx, service, actor, start := travelFixture(t)
	target := locationAt(t, ctx, s, "dromund_kaas", 10)
	other := planetID(t, ctx, s, "korriban")
	begin := make(chan struct{})
	done := make(chan error, 2)
	go func() {
		<-begin
		_, err := service.Move(ctx, actor, world.MoveInput{CharacterID: 1, LocationID: target, FromLocationID: start})
		done <- err
	}()
	go func() {
		<-begin
		_, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: other, FromLocationID: start})
		done <- err
	}()
	close(begin)
	successes := 0
	for range 2 {
		err := <-done
		if err == nil {
			successes++
		} else if !errors.Is(err, world.ErrPositionChanged) && !errors.Is(err, world.ErrWrongPlanet) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("simultaneous moves succeeded: %d", successes)
	}
}

func TestWorldMoveRechecksPlanetAndExpiry(t *testing.T) {
	for _, change := range []string{"control", "expiry"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, service, actor, start := travelFixture(t)
			planet := emptyNeutralPlanet(t, ctx, s)
			mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'arrival','Arrival',1,1),($1,'deep','Deep',2,46)`, planet)
			mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT faction_id FROM characters WHERE id=1) WHERE id=$1`, planet)
			moved, err := service.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: planet, FromLocationID: start})
			if err != nil {
				t.Fatal(err)
			}
			target := locationAt(t, ctx, s, "test_empty_neutral", 2)
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			var pid int
			if err = tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM planets WHERE id=$1 FOR UPDATE`, planet).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			var result *world.TravelResult
			go func() {
				var e error
				result, e = service.Move(ctx, actor, world.MoveInput{CharacterID: 1, LocationID: target, FromLocationID: moved.Position.Location.ID})
				done <- e
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			if change == "control" {
				_, err = tx.Exec(ctx, `UPDATE planets SET controlling_faction_id=NULL WHERE id=$1`, planet)
			} else {
				_, err = tx.Exec(ctx, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, actor.SessionHash)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			err = <-done
			if change == "control" {
				if err != nil || !result.ReturnedHome || result.Position.Planet.Code != "dromund_kaas" {
					t.Fatalf("return: %+v %v", result, err)
				}
			} else if !errors.Is(err, character.ErrUnauthenticated) {
				t.Fatalf("expired move: %v", err)
			}
		})
	}
}
