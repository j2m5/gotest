package db_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"gotest/internal/ability"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/wallet"
)

type executionFixture struct {
	s            *testStorage
	ctx          context.Context
	actor, other character.Actor
	in           combat.StartInput
	r            *db.CombatRepository
	battle       *combat.Battle
	abilityID    int
}

func newExecutionFixture(t *testing.T) executionFixture {
	t.Helper()
	s, ctx, actor, other, in := combatFixture(t)
	mustExec(t, ctx, s, `INSERT INTO character_abilities(character_id,ability_id)
 SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id ON CONFLICT DO NOTHING`)
	f := executionFixture{s: s, ctx: ctx, actor: actor, other: other, in: in, r: db.NewCombatRepository(s.pool)}
	if err := s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=$1`, in.CharacterID).Scan(&f.abilityID); err != nil {
		t.Fatal(err)
	}
	result, err := f.r.Start(ctx, actor, in)
	if err != nil {
		t.Fatal(err)
	}
	f.battle = result.Battle
	return f
}

func (f executionFixture) cast(key string) (*combat.Result, error) {
	return f.r.Use(f.ctx, f.actor, f.battle.ID, combat.UseInput{CharacterID: f.in.CharacterID, AbilityID: f.abilityID, Key: key})
}
func (f executionFixture) due(t *testing.T) {
	t.Helper()
	mustExec(t, f.ctx, f.s, `UPDATE battle_executions SET next_action_at=clock_timestamp() WHERE battle_id=$1`, f.battle.ID)
}
func (f executionFixture) view(t *testing.T) *combat.View {
	t.Helper()
	v, err := f.r.View(f.ctx, f.actor, f.battle.ID, f.in.CharacterID, 0)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCombatExecutionStartUseReplayAndHistory(t *testing.T) {
	f := newExecutionFixture(t)
	v := f.view(t)
	if len(v.Events) != 1 || v.Events[0].Kind != "started" || v.NextActionAt == nil || !v.ExpiresAt.After(f.battle.CreatedAt) {
		t.Fatalf("start %+v", v)
	}
	repeat, err := f.r.Start(f.ctx, f.actor, f.in)
	if err != nil || !repeat.Replayed || repeat.Battle.ID != f.battle.ID {
		t.Fatal(repeat, err)
	}
	use, err := f.cast("cast:one")
	if err != nil || use.Replayed || use.Battle.NPC.HPCurrent >= f.battle.NPC.HP || use.Character.MPCurrent != 90 {
		t.Fatal(use, err)
	}
	before := characterStatsSnapshot(t, f.ctx, f.s)
	again, err := f.cast("cast:one")
	if err != nil || !again.Replayed || again.EventSequence != use.EventSequence || characterStatsSnapshot(t, f.ctx, f.s) != before {
		t.Fatal("cast replay mutated resources", again, err)
	}
	_, err = f.r.Use(f.ctx, f.actor, f.battle.ID, combat.UseInput{CharacterID: f.in.CharacterID, AbilityID: f.abilityID + 1, Key: "cast:one"})
	if !errors.Is(err, combat.ErrConflict) {
		t.Fatal("key conflict", err)
	}
	_, err = f.cast("cast:two")
	if !errors.Is(err, ability.ErrGlobalCooldown) && !errors.Is(err, ability.ErrCooldown) {
		t.Fatal("cooldown", err)
	}
	v = f.view(t)
	if len(v.Events) != 2 || v.Events[1].Kind != "player_ability" || v.Events[1].Damage != f.battle.NPC.HP-use.Battle.NPC.HPCurrent {
		t.Fatalf("history %+v", v.Events)
	}
	page, err := f.r.View(f.ctx, f.actor, f.battle.ID, f.in.CharacterID, 1)
	if err != nil || len(page.Events) != 1 || page.Events[0].Sequence != 2 {
		t.Fatal(page, err)
	}
	if _, err = f.r.View(f.ctx, f.other, f.battle.ID, 3, 0); !errors.Is(err, combat.ErrNotFound) {
		t.Fatal("foreign history", err)
	}
}

func TestCombatExecutionDeniedCastStillCommitsDueAttack(t *testing.T) {
	f := newExecutionFixture(t)
	if _, err := f.cast("first"); err != nil {
		t.Fatal(err)
	}
	f.due(t)
	_, err := f.cast("cooldown")
	if !errors.Is(err, ability.ErrGlobalCooldown) && !errors.Is(err, ability.ErrCooldown) {
		t.Fatal(err)
	}
	v := f.view(t)
	if len(v.Events) != 3 || v.Events[2].Kind != "npc_ability" || v.Character.HPCurrent >= 100 {
		t.Fatal("owed attack rolled back", v.Events, v.Character.HPCurrent)
	}
	var rawMP int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT mp_current FROM characters WHERE id=$1`, f.in.CharacterID).Scan(&rawMP); err != nil || rawMP != 90 {
		t.Fatal("failed cast spent energy", rawMP, err)
	}
}

func TestCombatExecutionVictoryRewardsOnceAndRespawns(t *testing.T) {
	f := newExecutionFixture(t)
	mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
	result, err := f.cast("win")
	if err != nil || result.Battle.Status != combat.Won || result.Battle.NPC.HPCurrent != 0 {
		t.Fatal(result, err)
	}
	rules, _ := combat.RulesFor(f.battle.NPC.Level, f.battle.NPC.Rank)
	var credits, coins, xp, entries int64
	var respawning bool
	if err = f.s.pool.QueryRow(f.ctx, `SELECT
 (SELECT balance FROM wallets WHERE character_id=$2 AND currency='credits'),
 (SELECT balance FROM wallets WHERE user_id=$1 AND currency='cartel_coins'),
 (SELECT experience FROM characters WHERE id=$2),
 (SELECT count(*) FROM wallet_entries WHERE source='pve'),
 (SELECT respawn_at>clock_timestamp() FROM npc_spawns WHERE id=$3)`, f.actor.UserID, f.in.CharacterID, f.in.NPCSpawnID).Scan(&credits, &coins, &xp, &entries, &respawning); err != nil {
		t.Fatal(err)
	}
	if credits != rules.Credits || coins != 0 || xp != rules.Experience || entries != 1 || !respawning {
		t.Fatal(credits, coins, xp, entries, respawning)
	}
	before := characterStatsSnapshot(t, f.ctx, f.s)
	again, err := f.cast("win")
	if err != nil || !again.Replayed {
		t.Fatal(again, err)
	}
	ended, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID)
	if err != nil || !ended.Replayed || ended.Battle.Status != combat.Won || before != characterStatsSnapshot(t, f.ctx, f.s) {
		t.Fatal("late retreat changed win", ended, err)
	}
	v := f.view(t)
	if len(v.Events) != 3 || v.Events[2].Kind != "won" || v.Events[2].Experience != rules.Experience || v.Events[2].Credits != rules.Credits {
		t.Fatal(v.Events)
	}
	_, err = f.r.Start(f.ctx, f.actor, combat.StartInput{CharacterID: f.in.CharacterID, LocationID: f.in.LocationID, NPCSpawnID: f.in.NPCSpawnID, Key: "too-soon"})
	if !errors.Is(err, combat.ErrNPCUnavailable) {
		t.Fatal("respawning NPC available", err)
	}
}

func TestCombatExecutionRewardFailureRollsBackHitAndCooldown(t *testing.T) {
	f := newExecutionFixture(t)
	mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
	mustExec(t, f.ctx, f.s, `UPDATE wallets SET balance=$1 WHERE character_id=$2 AND currency='credits'`, int64(math.MaxInt64), f.in.CharacterID)
	before := characterStatsSnapshot(t, f.ctx, f.s)
	_, err := f.cast("overflow")
	if !errors.Is(err, wallet.ErrOverflow) {
		t.Fatal(err)
	}
	v := f.view(t)
	if before != characterStatsSnapshot(t, f.ctx, f.s) || len(v.Events) != 1 || v.Battle.Status != combat.Active || v.Battle.NPC.HPCurrent != v.Battle.NPC.HP {
		t.Fatal("partial victory", v)
	}
}

func TestCombatExecutionWorkersContinueAfterLogoutAndDoNotDuplicate(t *testing.T) {
	f := newExecutionFixture(t)
	f.due(t)
	mustExec(t, f.ctx, f.s, `DELETE FROM sessions WHERE user_id=$1`, f.actor.UserID)
	mustExec(t, f.ctx, f.s, `UPDATE users SET character_id=NULL WHERE id=$1`, f.actor.UserID)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Go(func() { _, err := db.NewCombatRepository(f.s.pool).ProcessDue(f.ctx); errs <- err })
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count, step int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT (SELECT count(*) FROM battle_events WHERE battle_id=$1),pattern_step FROM battle_executions WHERE battle_id=$1`, f.battle.ID).Scan(&count, &step); err != nil || count != 2 || step != 1 {
		t.Fatal(count, step, err)
	}
}

func TestCombatExecutionDueLethalHitPrecedesRetreat(t *testing.T) {
	f := newExecutionFixture(t)
	mustExec(t, f.ctx, f.s, `UPDATE characters SET hp_current=1 WHERE id=$1`, f.in.CharacterID)
	f.due(t)
	result, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID)
	if err != nil || result.Battle.Status != combat.Defeated || result.Character.HPCurrent != 0 {
		t.Fatal(result, err)
	}
	v := f.view(t)
	if len(v.Events) != 3 || v.Events[1].Kind != "npc_ability" || v.Events[2].Kind != "defeated" || v.NextActionAt != nil {
		t.Fatal(v.Events)
	}
	var ready bool
	if err := f.s.pool.QueryRow(f.ctx, `SELECT respawn_at IS NULL FROM npc_spawns WHERE id=$1`, f.in.NPCSpawnID).Scan(&ready); err != nil || !ready {
		t.Fatal(ready, err)
	}
}

func TestCombatExecutionRetreatPreservesResourcesAndOldCommands(t *testing.T) {
	f := newExecutionFixture(t)
	if _, err := f.cast("hit"); err != nil {
		t.Fatal(err)
	}
	before := characterStatsSnapshot(t, f.ctx, f.s)
	result, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID)
	if err != nil || result.Battle.Status != combat.Retreated || before != characterStatsSnapshot(t, f.ctx, f.s) {
		t.Fatal(result, err)
	}
	nextInput := f.in
	nextInput.Key = "next"
	next, err := f.r.Start(f.ctx, f.actor, nextInput)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.cast("late"); !errors.Is(err, combat.ErrFinished) {
		t.Fatal("old cast", err)
	}
	if _, err = f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID); err != nil {
		t.Fatal(err)
	}
	current, err := f.r.Current(f.ctx, f.actor, f.in.CharacterID)
	if err != nil || current.Battle.ID != next.Battle.ID {
		t.Fatal("new encounter affected", current, err)
	}
	if n := len(f.view(t).Events); n != 3 {
		t.Fatal(fmt.Sprint("old log changed: ", n))
	}
}
