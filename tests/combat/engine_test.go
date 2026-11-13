package combat_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/combat"
)

func engineFrame(t *testing.T) combat.Frame {
	t.Helper()
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	rules, err := combat.RulesFor(1, "normal")
	if err != nil {
		t.Fatal(err)
	}
	return combat.Frame{Battle: combat.Battle{ID: 1, CharacterID: 1, Status: combat.Active, CreatedAt: at,
		NPC: combat.Opponent{HP: rules.HP, HPCurrent: rules.HP}}, Rules: rules,
		Resources:    character.Resources{Health: character.Resource{Current: 100, Maximum: 100, UpdatedAt: at}, Energy: character.Resource{Current: 70, Maximum: 100, UpdatedAt: at}},
		NextActionAt: at.Add(2 * time.Second), ExpiresAt: at.Add(10 * time.Minute)}
}

func TestCombatRulesAndOverflowSafeDamage(t *testing.T) {
	for _, rank := range []string{"normal", "strong", "elite"} {
		low, err := combat.RulesFor(1, rank)
		if err != nil {
			t.Fatal(err)
		}
		high, err := combat.RulesFor(50, rank)
		if err != nil || high.HP <= low.HP || high.Experience <= low.Experience || high.Credits <= low.Credits {
			t.Fatalf("scaling: %+v %+v %v", low, high, err)
		}
		if err := high.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		level int
		rank  string
	}{{0, "normal"}, {51, "normal"}, {1, "boss"}} {
		if _, err := combat.RulesFor(tc.level, tc.rank); err == nil {
			t.Fatal("invalid balance accepted", tc)
		}
	}
	if got := combat.Damage(15, 10, 10, 0); got != 18 {
		t.Fatal(got)
	}
	if got := combat.Damage(0, 100, 100, 0); got != 0 {
		t.Fatal("a heal must not gain implicit damage", got)
	}
	if got := combat.Damage(math.MaxInt32, math.MaxInt32, math.MaxInt32, 0); got != math.MaxInt32 {
		t.Fatal("overflow", got)
	}
	if got := combat.Damage(1, 0, 0, math.MaxInt32); got != 1 {
		t.Fatal("positive hit minimum", got)
	}
}

func TestCombatCatchUpMatchesIndividualTicksAndPreservesRegeneration(t *testing.T) {
	f := engineFrame(t)
	next, events, err := combat.Advance(f, f.Battle.CreatedAt.Add(7*time.Second))
	if err != nil || len(events) != 3 {
		t.Fatalf("events: %+v %v", events, err)
	}
	individual := f
	for i, sec := range []int{2, 4, 7} {
		var one []combat.Event
		individual, one, err = combat.Advance(individual, f.Battle.CreatedAt.Add(time.Duration(sec)*time.Second))
		if err != nil || len(one) != 1 || !reflect.DeepEqual(one[0], events[i]) {
			t.Fatalf("tick %d: %+v %v", i, one, err)
		}
	}
	if !reflect.DeepEqual(next, individual) {
		t.Fatalf("catch-up differs: %+v %+v", next, individual)
	}
	// 6 + 6 + 12 damage, plus one HP and 2/3 HP recovered after the first hit.
	if next.Resources.Health.Current != 77 || next.Resources.Health.Remainder != 200000 {
		t.Fatalf("lost fractional regen: %+v", next.Resources.Health)
	}
	if next.Resources.Energy != f.Resources.Energy {
		t.Fatal("NPC attack changed energy")
	}
	if next.PatternStep != 3 || !next.NextActionAt.Equal(f.Battle.CreatedAt.Add(9*time.Second)) {
		t.Fatal("pattern did not advance", next)
	}
	for i, sec := range []int{2, 4, 7} {
		if events[i].Sequence != i+1 || !events[i].OccurredAt.Equal(f.Battle.CreatedAt.Add(time.Duration(sec)*time.Second)) {
			t.Fatal(events)
		}
	}
}

func TestCombatDueBoundaryDeathAndTimeout(t *testing.T) {
	f := engineFrame(t)
	_, events, err := combat.Advance(f, f.NextActionAt.Add(-time.Millisecond))
	if err != nil || len(events) != 0 {
		t.Fatal(events, err)
	}
	f.Resources.Health.Current = 1
	next, events, err := combat.Advance(f, f.Battle.CreatedAt.Add(time.Hour))
	if err != nil || next.Battle.Status != combat.Defeated || len(events) != 2 || next.Resources.Health.Current != 0 || !next.Resources.Health.UpdatedAt.Equal(f.NextActionAt) {
		t.Fatalf("death: %+v %+v %v", next, events, err)
	}
	if next.Battle.NPC.HPCurrent != next.Battle.NPC.HP || !next.NextActionAt.IsZero() {
		t.Fatal("NPC not released")
	}
	f = engineFrame(t)
	f.Defense = math.MaxInt32
	next, events, err = combat.Advance(f, f.Battle.CreatedAt.Add(24*time.Hour))
	if err != nil || next.Battle.Status != combat.Expired || len(events) > 601 || !next.Battle.EndedAt.Equal(f.ExpiresAt) {
		t.Fatalf("bounded timeout: %+v events %d %v", next, len(events), err)
	}
	again, replayEvents, err := combat.Advance(next, f.Battle.CreatedAt.Add(48*time.Hour))
	if err != nil || len(replayEvents) != 0 || !reflect.DeepEqual(again, next) {
		t.Fatal("terminal replay changed result")
	}
	f = engineFrame(t)
	f.ExpiresAt = f.NextActionAt
	_, events, err = combat.Advance(f, f.NextActionAt)
	if err != nil || len(events) != 1 || events[0].Kind != "expired" {
		t.Fatal("timeout must win tie", events, err)
	}
}

func TestCombatPlayerEffectClampsHealingAndOverkill(t *testing.T) {
	f := engineFrame(t)
	f.Resources.Health.Current = 95
	f.Battle.NPC.HPCurrent = 3
	next, event, err := combat.PlayerEffect(f, 15, 25, 10, 10, "test", "Тест", f.Battle.CreatedAt)
	if err != nil || event.Damage != 3 || event.Healing != 5 || next.Resources.Health.Current != 100 || next.Battle.NPC.HPCurrent != 0 {
		t.Fatal(next, event, err)
	}
	if next.Battle.Status != combat.Won {
		t.Fatal("lethal player effect did not win")
	}
}
