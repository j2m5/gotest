package db_test

import (
	"errors"
	"reflect"
	"testing"

	"gotest/internal/db"
	"gotest/internal/npc"
	"gotest/internal/world"
)

func TestPopulatedWorldTravelAndControl(t *testing.T) {
	s, ctx, travel, actor, _ := travelFixture(t)
	enemies := npc.NewService(db.NewNPCRepository(s.pool))
	for _, tc := range []struct {
		planet  string
		ownerID int
		faction string
		core    bool
	}{
		{"arkania", 2, "galactic_republic", false},
		{"balmorra", 2, "galactic_republic", false},
		{"ord_mantell", 2, "galactic_republic", false},
		{"athiss", 1, "sith_empire", false},
		{"thule", 1, "sith_empire", false},
		{"ossus", 1, "sith_empire", false},
		{"carida", 2, "galactic_republic", false},
		{"cato_neimoidia", 2, "galactic_republic", false},
		{"shili", 2, "galactic_republic", false},
		{"ord_radama", 1, "sith_empire", false},
		{"mirial", 1, "sith_empire", false},
		{"telos_iv", 1, "sith_empire", false},
		{"commenor", 2, "galactic_republic", false},
		{"ralltiir", 2, "galactic_republic", false},
		{"humbarine", 2, "galactic_republic", false},
		{"neimoidia", 2, "galactic_republic", false},
		{"dathomir", 1, "sith_empire", false},
		{"saleucami", 1, "sith_empire", false},
		{"sriluur", 1, "sith_empire", false},
		{"dac", 1, "sith_empire", false},
		{"corellia", 2, "galactic_republic", true},
		{"alderaan", 2, "galactic_republic", true},
		{"kuat", 2, "galactic_republic", true},
		{"ziost", 1, "sith_empire", true},
		{"begeren", 1, "sith_empire", true},
		{"bosthirda", 1, "sith_empire", true},
		{"chandrila", 2, "galactic_republic", true},
		{"duro", 2, "galactic_republic", true},
		{"brentaal_iv", 2, "galactic_republic", true},
		{"jaguada", 1, "sith_empire", true},
		{"rhelg", 1, "sith_empire", true},
		{"khar_delba", 1, "sith_empire", true},
		{"anaxes", 2, "galactic_republic", true},
		{"alsakan", 2, "galactic_republic", true},
		{"rendili", 2, "galactic_republic", true},
		{"dromund_fels", 1, "sith_empire", true},
		{"krayiss_ii", 1, "sith_empire", true},
		{"chhodos", 1, "sith_empire", true},
		{"caamas", 2, "galactic_republic", true},
		{"esseles", 2, "galactic_republic", true},
		{"rhinnal", 2, "galactic_republic", true},
		{"corulag", 2, "galactic_republic", true},
		{"ashas_ree", 1, "sith_empire", true},
		{"kalsunor", 1, "sith_empire", true},
		{"korriz", 1, "sith_empire", true},
		{"upekzar", 1, "sith_empire", true},
	} {
		t.Run(tc.planet, func(t *testing.T) {
			planet := planetID(t, ctx, s, tc.planet)
			mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, tc.ownerID)
			nav, err := travel.Navigation(ctx, actor, tc.ownerID)
			if err != nil {
				t.Fatal(err)
			}
			moved, err := travel.Travel(ctx, actor, world.TravelInput{CharacterID: tc.ownerID, PlanetID: planet, FromLocationID: nav.Position.Location.ID})
			if err != nil {
				t.Fatalf("owned populated world unavailable: %v", err)
			}
			p := moved.Position.Planet
			if p.ControllingFaction == nil || p.ControllingFaction.Code != tc.faction || moved.Position.Location.RecommendedLevel != 1 {
				t.Fatalf("wrong ownership/arrival: %+v", moved.Position)
			}
			if tc.core {
				if p.CoreFaction == nil || p.CoreFaction.Code != tc.faction {
					t.Fatalf("wrong protected core: %+v", p)
				}
			} else if p.CoreFaction != nil {
				t.Fatalf("frontier world became protected: %+v", p)
			}
			areas, err := travel.Locations(ctx, actor, tc.ownerID, planet)
			if err != nil || len(areas.Locations) != 10 {
				t.Fatalf("populated world locations: %+v %v", areas, err)
			}
			last := areas.Locations[9]
			if last.RecommendedLevel != 46 {
				t.Fatalf("last recommendation: %d", last.RecommendedLevel)
			}
			if _, err = travel.Move(ctx, actor, world.MoveInput{CharacterID: tc.ownerID, LocationID: last.ID, FromLocationID: moved.Position.Location.ID}); err != nil {
				t.Fatal(err)
			}
			list, err := enemies.List(ctx, actor, tc.ownerID, last.ID)
			if err != nil {
				t.Fatal(err)
			}
			var levels []int
			for _, enemy := range list.NPCs {
				levels = append(levels, enemy.Level)
			}
			if !reflect.DeepEqual(levels, []int{46, 46, 48, 50}) {
				t.Fatalf("final NPC levels: %v", levels)
			}
			otherID := 3 - tc.ownerID
			mustExec(t, ctx, s, `UPDATE users SET character_id=$1 WHERE id=1`, otherID)
			nav, err = travel.Navigation(ctx, actor, otherID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = travel.Travel(ctx, actor, world.TravelInput{CharacterID: otherID, PlanetID: planet, FromLocationID: nav.Position.Location.ID}); !errors.Is(err, world.ErrAccessDenied) {
				t.Fatalf("opposing faction allowed: %v", err)
			}
		})
	}
}
