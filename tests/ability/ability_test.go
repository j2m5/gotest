package ability_test

import (
	"testing"
	"time"

	"gotest/internal/ability"
)

func TestAbilityAvailabilityUsesServerDeadlinesAndFullCost(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	global := at.Add(1500 * time.Millisecond)
	personal := at.Add(10 * time.Second)
	definition := ability.Definition{MPCost: 10}
	for _, tc := range []struct {
		name       string
		ready, gcd *time.Time
		hp, mp     int
		now        time.Time
		want       ability.Availability
	}{
		{"before GCD boundary", nil, &global, 100, 10, at.Add(1499 * time.Millisecond), ability.GlobalCooldown},
		{"at GCD boundary", nil, &global, 100, 10, global, ability.Ready},
		{"personal persists", &personal, &global, 100, 10, global, ability.Cooldown},
		{"at personal boundary", &personal, &global, 100, 10, personal, ability.Ready},
		{"cannot pay partial cost", nil, nil, 100, 9, at, ability.InsufficientEnergy},
		{"defeated", nil, nil, 0, 100, at, ability.Defeated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ability.AvailabilityAt(definition, tc.ready, tc.gcd, tc.hp, tc.mp, tc.now); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
