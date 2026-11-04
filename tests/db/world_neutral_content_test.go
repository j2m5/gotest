package db_test

import (
	"errors"
	"reflect"
	"testing"

	"gotest/internal/db"
	"gotest/internal/npc"
	"gotest/internal/world"
)

var neutralContentPlanets = []string{
	"belsavis", "bothawui", "dantooine", "taris", "hoth", "tatooine", "manaan", "onderon",
	"makeb", "nal_hutta", "kashyyyk", "trandosha", "ryloth", "rodia", "sullust", "ithor",
	"ilum", "rishi", "mandalore", "concord_dawn", "cathar", "myrkr", "iridonia", "ambria",
	"sleheyron", "toydaria", "kwenn", "nar_kreeta", "klatooine", "vodran", "kintan", "ubrikkia",
	"dorin", "glee_anselm", "gand", "togoria", "eres_iii", "taanab", "zeltros", "cerea", "bimmisaari",
	"malastare", "kessel", "eriadu", "sluis_van", "fondor", "bestine_iv", "denon", "mygeeto", "muunilinst",
}

func TestPopulatedNeutralWorldTravelAndControl(t *testing.T) {
	s, ctx, travel, actor, _ := travelFixture(t)
	enemies := npc.NewService(db.NewNPCRepository(s.pool))
	for _, code := range neutralContentPlanets {
		t.Run(code, func(t *testing.T) {
			planet := planetID(t, ctx, s, code)
			var neutral bool
			if err := s.pool.QueryRow(ctx, `SELECT core_faction_id IS NULL AND controlling_faction_id IS NULL FROM planets WHERE id=$1`, planet).Scan(&neutral); err != nil || !neutral {
				t.Fatalf("populated world must remain neutral: %v %v", neutral, err)
			}
			for _, characterID := range []int{1, 2} {
				mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, characterID)
				nav, err := travel.Navigation(ctx, actor, characterID)
				if err != nil {
					t.Fatal(err)
				}
				var destination *world.Destination
				for _, d := range nav.Destinations {
					if d.PlanetID == planet {
						destination = &d
					}
				}
				if destination == nil || destination.CanTravel || destination.Reason != "invasion_required" {
					t.Fatalf("neutral destination: %+v", destination)
				}
				if _, err = travel.Travel(ctx, actor, world.TravelInput{CharacterID: characterID, PlanetID: planet, FromLocationID: nav.Position.Location.ID}); !errors.Is(err, world.ErrAccessDenied) {
					t.Fatalf("neutral world allowed: %v", err)
				}
				after, err := travel.Navigation(ctx, actor, characterID)
				if err != nil || !reflect.DeepEqual(nav.Position, after.Position) {
					t.Fatalf("rejected flight changed position: %+v %v", after, err)
				}
			}
			var sharedIDs []int
			for _, ownerID := range []int{1, 2} {
				// Ownership changes are fixtures in this isolated database only.
				mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=(SELECT faction_id FROM characters WHERE id=$1) WHERE id=$2`, ownerID, planet)
				mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, ownerID)
				nav, err := travel.Navigation(ctx, actor, ownerID)
				if err != nil {
					t.Fatal(err)
				}
				moved, err := travel.Travel(ctx, actor, world.TravelInput{CharacterID: ownerID, PlanetID: planet, FromLocationID: nav.Position.Location.ID})
				if err != nil {
					t.Fatalf("owned populated world unavailable: %v", err)
				}
				faction, home := "sith_empire", "dromund_kaas"
				if ownerID == 2 {
					faction, home = "galactic_republic", "coruscant"
				}
				p := moved.Position.Planet
				if p.CoreFaction != nil || p.ControllingFaction == nil || p.ControllingFaction.Code != faction || moved.Position.Location.RecommendedLevel != 1 {
					t.Fatalf("wrong owner/arrival: %+v", moved.Position)
				}
				areas, err := travel.Locations(ctx, actor, ownerID, planet)
				if err != nil || len(areas.Locations) != 10 {
					t.Fatalf("locations: %+v %v", areas, err)
				}
				last := areas.Locations[9]
				if last.RecommendedLevel != 46 || areas.Locations[0].ID != moved.Position.Location.ID {
					t.Fatalf("wrong entry or final recommendation: %+v", areas)
				}
				if _, err = travel.Move(ctx, actor, world.MoveInput{CharacterID: ownerID, LocationID: last.ID, FromLocationID: moved.Position.Location.ID}); err != nil {
					t.Fatal(err)
				}
				list, err := enemies.List(ctx, actor, ownerID, last.ID)
				if err != nil {
					t.Fatal(err)
				}
				var ids, levels []int
				for _, enemy := range list.NPCs {
					ids = append(ids, enemy.ID)
					levels = append(levels, enemy.Level)
				}
				if !reflect.DeepEqual(levels, []int{46, 46, 48, 50}) {
					t.Fatalf("final NPC levels: %v", levels)
				}
				if sharedIDs != nil && !reflect.DeepEqual(sharedIDs, ids) {
					t.Fatalf("ownership changed NPC instances: %v -> %v", sharedIDs, ids)
				}
				sharedIDs = ids
				otherID := 3 - ownerID
				mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, otherID)
				nav, err = travel.Navigation(ctx, actor, otherID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = travel.Travel(ctx, actor, world.TravelInput{CharacterID: otherID, PlanetID: planet, FromLocationID: nav.Position.Location.ID}); !errors.Is(err, world.ErrAccessDenied) {
					t.Fatalf("opposing faction allowed: %v", err)
				}
				mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, ownerID)
				mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=NULL WHERE id=$1`, planet)
				nav, err = travel.Navigation(ctx, actor, ownerID)
				if err != nil || !nav.ReturnRequired || nav.Position.Location.ID != last.ID {
					t.Fatalf("lost world read changed position: %+v %v", nav, err)
				}
				if _, err = enemies.List(ctx, actor, ownerID, last.ID); !errors.Is(err, world.ErrAccessDenied) {
					t.Fatalf("lost-world NPCs still accessible: %v", err)
				}
				returned, err := travel.Move(ctx, actor, world.MoveInput{CharacterID: ownerID, LocationID: areas.Locations[0].ID, FromLocationID: last.ID})
				if err != nil || !returned.ReturnedHome || returned.Position.Planet.Code != home {
					t.Fatalf("lost world return: %+v %v", returned, err)
				}
			}
		})
	}
}
