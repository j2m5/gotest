package db_test

import (
	"errors"
	"reflect"
	"testing"

	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/npc"
	"gotest/internal/world"
)

func TestNPCInstancesSharedAndReadsDoNotMutate(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	a := repositoryActor(t, ctx, s, 1, "npc-first")
	b := repositoryActor(t, ctx, s, 2, "npc-second")
	mustExec(t, ctx, s, `UPDATE users SET character_id=CASE id WHEN 1 THEN 2 WHEN 2 THEN 3 END WHERE id IN (1,2)`)
	location := locationAt(t, ctx, s, "coruscant", 1)
	service := npc.NewService(db.NewNPCRepository(s.pool))
	snapshot := `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM npc_spawns t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM characters t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM sessions t))::text`
	var before, after string
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	first, err := service.List(ctx, a, 2, location)
	if err != nil || first.LocationID != location || len(first.NPCs) != 4 {
		t.Fatalf("first list: %+v %v", first, err)
	}
	second, err := service.List(ctx, b, 3, location)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("players see different instances: %+v %+v %v", first, second, err)
	}
	repeat, err := service.List(ctx, a, 2, location)
	if err != nil || !reflect.DeepEqual(first, repeat) {
		t.Fatalf("read changed pool: %+v %v", repeat, err)
	}
	if err = s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
		t.Fatalf("NPC GET mutated state: %v", err)
	}
	next := locationAt(t, ctx, s, "coruscant", 10)
	if _, err = world.NewService(db.NewWorldRepository(s.pool)).Move(ctx, a, world.MoveInput{CharacterID: 2, LocationID: next, FromLocationID: location}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.List(ctx, a, 2, location); !errors.Is(err, world.ErrPositionChanged) {
		t.Fatalf("old location visible: %v", err)
	}
	other, err := service.List(ctx, b, 3, location)
	if err != nil || !reflect.DeepEqual(first, other) {
		t.Fatalf("movement changed another player's NPC pool: %+v %v", other, err)
	}
	high, err := service.List(ctx, a, 2, next)
	if err != nil || len(high.NPCs) != 4 || high.NPCs[3].Level != 50 {
		t.Fatalf("level50 enemies: %+v %v", high, err)
	}
}

func TestNPCListRequiresCurrentActorLocationAndControl(t *testing.T) {
	for _, change := range []string{"selection", "no_selection", "wrong_session", "revoked", "expired", "ban", "location", "control"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, travel, actor, start := travelFixture(t)
			service := npc.NewService(db.NewNPCRepository(s.pool))
			want := character.ErrUnauthenticated
			switch change {
			case "selection":
				mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1`)
				want = world.ErrCharacterChanged
			case "no_selection":
				mustExec(t, ctx, s, `UPDATE users SET character_id=NULL WHERE id=1`)
				want = world.ErrCharacterRequired
			case "wrong_session":
				actor.UserID = 2
			case "revoked":
				mustExec(t, ctx, s, `DELETE FROM sessions WHERE user_id=1`)
			case "expired":
				mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
			case "ban":
				mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`)
			case "location":
				start = locationAt(t, ctx, s, "dromund_kaas", 10)
				want = world.ErrPositionChanged
			case "control":
				planet := emptyNeutralPlanet(t, ctx, s)
				mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'test_area','Area',1,1);`, planet)
				mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT faction_id FROM characters WHERE id=1) WHERE id=$1`, planet)
				moved, err := travel.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: planet, FromLocationID: start})
				if err != nil {
					t.Fatal(err)
				}
				start = moved.Position.Location.ID
				mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=NULL WHERE id=$1`, planet)
				want = world.ErrAccessDenied
			}
			if _, err := service.List(ctx, actor, 1, start); !errors.Is(err, want) {
				t.Fatalf("%s: %v want %v", change, err, want)
			}
		})
	}
}

func TestNPCUnpopulatedLocationIsEmpty(t *testing.T) {
	s, ctx, travel, actor, start := travelFixture(t)
	planet := emptyCorePlanet(t, ctx, s)
	mustExec(t, ctx, s, `INSERT INTO locations(planet_id,code,name,position,recommended_level) VALUES($1,'empty','Empty test area',1,1)`, planet)
	result, err := travel.Travel(ctx, actor, world.TravelInput{CharacterID: 1, PlanetID: planet, FromLocationID: start})
	if err != nil {
		t.Fatal(err)
	}
	list, err := npc.NewService(db.NewNPCRepository(s.pool)).List(ctx, actor, 1, result.Position.Location.ID)
	if err != nil || list.NPCs == nil || len(list.NPCs) != 0 {
		t.Fatalf("empty pool: %+v %v", list, err)
	}
}
