package routes_test

import (
	"context"
	"testing"
	"time"

	"gotest/internal/ability"
	"gotest/internal/character"
)

type abilityStub struct {
	empty, filled bool
	err           error
}

func (s *abilityStub) List(_ context.Context, _ character.Actor, _ int) (*ability.Loadout, error) {
	c := (&contractCharacters{}).sample()
	result := &ability.Loadout{Character: c, ObservedAt: c.Regeneration.ObservedAt,
		GlobalCooldownMS: ability.GlobalCooldownDuration.Milliseconds(), Abilities: []ability.Owned{}}
	if !s.empty {
		owned := ability.Owned{Definition: ability.Definition{ID: 1, Code: "basic_attack", Name: "Удар", Description: "Наносит базовый урон", Damage: 15, MPCost: 10, CooldownMS: 1500}, Availability: ability.Ready}
		if s.filled {
			owned.Icon = new("images/abilities/hammershot.png")
			owned.ReadyAt = new(result.ObservedAt.Add(time.Second))
			result.GlobalReadyAt = owned.ReadyAt
			owned.Availability = ability.Cooldown
		}
		result.Abilities = append(result.Abilities, owned)
	}
	return result, s.err
}

func TestAbilitiesHaveNoPublicCastEndpoint(t *testing.T) {
	tc := contractCase{method: "POST", path: "/api/abilities", body: `{"character_id":10,"ability_id":1}`, status: 405}
	server := contractServer(t, tc)
	response, _ := contractExchange(t, server, contractRequest(t, server, tc), tc.status)
	if response.Header.Get("Allow") != "GET" {
		t.Fatalf("unexpected methods: %s", response.Header.Get("Allow"))
	}
}
