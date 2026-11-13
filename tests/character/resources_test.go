package character_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"gotest/internal/character"
)

func TestResourceRecoveryCarriesRemainder(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name             string
		current, maximum int
		carry            int64
		elapsed          time.Duration
		want             int
		wantCarry        int64
	}{
		{"odd capacity crosses next point", 1, 7, 290000, 2 * time.Second, 2, 4000},
		{"below next point", 50, 100, 100000, time.Second, 50, 200000},
		{"full discards carry", 99, 100, 250000, time.Second, 100, 0},
		{"full cannot bank time", 100, 100, 0, time.Hour, 100, 0},
		{"future anchor keeps carry", 10, 100, 200000, -time.Second, 10, 200000},
		{"zero capacity", 0, 0, 0, time.Hour, 0, 0},
		{"int32 range", 0, math.MaxInt32, 150000, 150 * time.Second, 1073741824, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, carry := character.RecoverResourceWithRemainder(tc.current, tc.maximum, tc.carry, at, at.Add(tc.elapsed), 5*time.Minute)
			if got != tc.want || carry != tc.wantCarry {
				t.Fatalf("got %d + %d/300000; want %d + %d/300000", got, carry, tc.want, tc.wantCarry)
			}
		})
	}
}

func TestResourceChangesPreserveRecoveryBetweenHits(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	state := character.Resources{
		Health: character.Resource{Current: 50, Maximum: 100, UpdatedAt: at},
		Energy: character.Resource{Current: 100, Maximum: 100, UpdatedAt: at},
	}
	for second := 1; second <= 3; second++ {
		var err error
		state, err = character.ChangeResources(state, character.ResourceChange{HP: -1}, at.Add(time.Duration(second)*time.Second))
		if err != nil {
			t.Fatal(err)
		}
	}
	if state.Health.Current != 48 || state.Health.Remainder != 0 {
		t.Fatalf("frequent damage lost regeneration: %+v", state.Health)
	}
	if !state.Energy.UpdatedAt.Equal(at) {
		t.Fatal("HP-only changes touched energy")
	}
}

func TestResourceChangesAtSameTimeAndBounds(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	state := character.Resources{
		Health: character.Resource{Current: 50, Maximum: 100, Remainder: 250000, UpdatedAt: at},
		Energy: character.Resource{Current: 10, Maximum: 100, Remainder: 50000, UpdatedAt: at},
	}
	changed, err := character.ChangeResources(state, character.ResourceChange{HP: -1, MP: -10}, at)
	if err != nil || changed.Health.Current != 49 || changed.Health.Remainder != 250000 || changed.Energy.Current != 0 || changed.Energy.Remainder != 50000 {
		t.Fatalf("same-time change: %+v %v", changed, err)
	}
	if _, err = character.ChangeResources(state, character.ResourceChange{MP: -11}, at); !errors.Is(err, character.ErrInsufficientEnergy) {
		t.Fatalf("overspend accepted: %v", err)
	}
	dead, err := character.ChangeResources(state, character.ResourceChange{HP: -500}, at)
	if err != nil || dead.Health.Current != 0 || dead.Health.Remainder != 0 {
		t.Fatalf("lethal damage: %+v %v", dead, err)
	}
	full, err := character.ChangeResources(state, character.ResourceChange{HP: math.MaxInt32, MP: math.MaxInt32}, at)
	if err != nil || full.Health.Current != 100 || full.Energy.Current != 100 || full.Health.Remainder != 0 || full.Energy.Remainder != 0 {
		t.Fatalf("full bounds: %+v %v", full, err)
	}
	if _, err = character.ChangeResources(state, character.ResourceChange{HP: math.MaxInt32 + 1}, at); !errors.Is(err, character.ErrInvalidResourceChange) {
		t.Fatalf("unbounded delta: %v", err)
	}
}

func TestResourceChangesDoNotBankFullOrFutureTime(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	full := character.Resources{Health: character.Resource{Current: 100, Maximum: 100, UpdatedAt: at}, Energy: character.Resource{Current: 0, Maximum: 0, UpdatedAt: at}}
	hit, err := character.ChangeResources(full, character.ResourceChange{HP: -10}, at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	current, carry := character.RecoverResourceWithRemainder(hit.Health.Current, 100, hit.Health.Remainder, hit.Health.UpdatedAt, at.Add(time.Hour+time.Second), 5*time.Minute)
	if current != 90 || carry != 100000 {
		t.Fatalf("banked full recovery: %d %d", current, carry)
	}
	future := full
	future.Health = character.Resource{Current: 50, Maximum: 100, Remainder: 200000, UpdatedAt: at.Add(time.Minute)}
	hit, err = character.ChangeResources(future, character.ResourceChange{HP: -1}, at)
	if err != nil || !hit.Health.UpdatedAt.Equal(future.Health.UpdatedAt) || hit.Health.Remainder != 200000 {
		t.Fatalf("future anchor rewound: %+v %v", hit, err)
	}
}
