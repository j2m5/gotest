package routes_test

import (
	"context"
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/combat"
)

type combatStub struct {
	empty bool
	err   error
}

func (s *combatStub) Current(_ context.Context, _ character.Actor, _ int) (*combat.State, error) {
	c := (&contractCharacters{}).sample()
	state := &combat.State{Character: c}
	if !s.empty {
		state.Battle = &combat.Battle{ID: 9007199254740993, CharacterID: c.ID, LocationID: c.LocationID, Status: combat.Active, CreatedAt: c.Regeneration.ObservedAt,
			NPC: combat.Opponent{SpawnID: 4, TemplateID: 2, Name: "Дроид", Level: 5, Rank: "strong", HP: 200, HPCurrent: 73}}
	}
	return state, s.err
}

func TestCombatCurrentEndpointIsReadOnly(t *testing.T) {
	tc := contractCase{method: "POST", path: "/api/battles/current", body: `{"character_id":10,"npc_spawn_id":1}`, status: 405}
	server := contractServer(t, tc)
	response, _ := contractExchange(t, server, contractRequest(t, server, tc), tc.status)
	if response.Header.Get("Allow") != "GET" {
		t.Fatalf("unexpected methods: %s", response.Header.Get("Allow"))
	}
}

func (s *combatStub) result(ctx context.Context, actor character.Actor) (*combat.Result, error) {
	state, _ := (&combatStub{}).Current(ctx, actor, 10)
	if s.empty {
		state.Battle.Status = combat.Won
		state.Battle.NPC.HPCurrent = 0
		state.Battle.EndedAt = &state.Battle.CreatedAt
	}
	return &combat.Result{State: *state, EventSequence: 2, Replayed: s.empty}, s.err
}
func (s *combatStub) Start(ctx context.Context, actor character.Actor, _ combat.StartInput) (*combat.Result, error) {
	return s.result(ctx, actor)
}
func (s *combatStub) Use(ctx context.Context, actor character.Actor, _ int64, _ combat.UseInput) (*combat.Result, error) {
	return s.result(ctx, actor)
}
func (s *combatStub) Retreat(ctx context.Context, actor character.Actor, _ int64, _ int) (*combat.Result, error) {
	return s.result(ctx, actor)
}
func (s *combatStub) View(ctx context.Context, actor character.Actor, _ int64, _ int, _ int) (*combat.View, error) {
	result, _ := s.result(ctx, actor)
	next := result.Battle.CreatedAt.Add(2 * time.Second)
	v := &combat.View{State: result.State, NextActionAt: &next, ExpiresAt: next.Add(10 * time.Minute), Events: []combat.Event{}}
	if s.empty {
		v.NextActionAt = nil
	} else {
		v.Events = append(v.Events, combat.Event{Sequence: 1, Kind: "player_ability", Actor: "player", AbilityCode: "strike", AbilityName: "Удар", Damage: 18, Healing: 5, CharacterHP: 100, CharacterMP: 40, NPCHP: 73, OccurredAt: result.Battle.CreatedAt})
		v.NextAfter = new(1)
	}
	return v, s.err
}

func TestCombatExecutionHTTP(t *testing.T) {
	for _, tc := range []contractCase{
		{method: "POST", path: "/api/battles/start", body: `{"character_id":10,"location_id":1,"npc_spawn_id":4,"request_key":"start:1"}`, status: 200},
		{method: "POST", path: "/api/battles/use", body: `{"battle_id":"9007199254740993","character_id":10,"ability_id":1,"request_key":"cast:1"}`, status: 200},
		{method: "POST", path: "/api/battles/retreat", body: `{"battle_id":"9007199254740993","character_id":10}`, status: 200},
		{method: "GET", path: "/api/battles/detail?battle_id=9007199254740993&character_id=10", status: 200},
		{method: "POST", path: "/api/battles/use", body: `{"battle_id":9007199254740993,"character_id":10,"ability_id":1,"request_key":"cast:1"}`, status: 400},
		{method: "POST", path: "/api/battles/use", body: `{"battle_id":"01","character_id":10,"ability_id":1,"request_key":"cast:1"}`, status: 400},
		{method: "GET", path: "/api/battles/detail?battle_id=1&character_id=10&after_sequence=", status: 400},
		{method: "POST", path: "/api/battles/start", body: `{"character_id":10,"location_id":1,"npc_spawn_id":4,"request_key":"start:1","hp":1}`, status: 400},
	} {
		t.Run(tc.method+tc.path+tc.body, func(t *testing.T) {
			server := contractServer(t, tc)
			contractExchange(t, server, contractRequest(t, server, tc), tc.status)
		})
	}
}
