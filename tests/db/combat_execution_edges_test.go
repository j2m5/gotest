package db_test

import (
	"encoding/json"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
)

func TestCombatExecutionConcurrentVictoryAndWorker(t *testing.T) {
	f := newExecutionFixture(t)
	mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
	f.due(t)
	start := make(chan struct{})
	results := make(chan *combat.Result, 2)
	errs := make(chan error, 3)
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() { <-start; result, err := f.cast("same-win"); results <- result; errs <- err })
	}
	wg.Go(func() { <-start; _, err := f.r.ProcessDue(f.ctx); errs <- err })
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	replays := 0
	for result := range results {
		if result.Replayed {
			replays++
		}
	}
	if replays != 1 {
		t.Fatal("duplicate accepted cast", replays)
	}
	v := f.view(t)
	if v.Battle.Status != combat.Won || len(v.Events) != 4 || v.Events[1].Kind != "npc_ability" || v.Events[2].Kind != "player_ability" || v.Events[3].Kind != "won" {
		t.Fatal("event ordering", v.Events)
	}
	var awards, mp int
	if err := f.s.pool.QueryRow(f.ctx, `SELECT (SELECT count(*) FROM wallet_entries WHERE source='pve'),mp_current FROM characters WHERE id=$1`, f.in.CharacterID).Scan(&awards, &mp); err != nil || awards != 1 || mp != 90 {
		t.Fatal(awards, mp, err)
	}
}

func TestCombatExecutionConcurrentHitRetreatAndNPCAttack(t *testing.T) {
	for _, lethal := range []bool{false, true} {
		t.Run(map[bool]string{false: "nonlethal NPC", true: "lethal NPC"}[lethal], func(t *testing.T) {
			f := newExecutionFixture(t)
			mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
			if lethal {
				mustExec(t, f.ctx, f.s, `UPDATE characters SET hp_current=1,hp_updated_at=clock_timestamp(),hp_regen_remainder=0 WHERE id=$1`, f.in.CharacterID)
			}
			f.due(t)
			start, done := make(chan struct{}), make(chan error, 3)
			go func() { <-start; _, err := f.cast("terminal-race"); done <- err }()
			go func() { <-start; _, err := f.r.Retreat(f.ctx, f.actor, f.battle.ID, f.in.CharacterID); done <- err }()
			go func() { <-start; _, err := f.r.ProcessDue(f.ctx); done <- err }()
			close(start)
			var raceError error
			for range 3 {
				if err := <-done; err != nil && !errors.Is(err, combat.ErrFinished) {
					raceError = err
				}
			}
			if raceError != nil {
				t.Fatal(raceError)
			}
			view := f.view(t)
			if lethal && view.Battle.Status != combat.Defeated {
				t.Fatal("a due lethal attack was bypassed", view.Battle.Status)
			}
			if !lethal && view.Battle.Status != combat.Won && view.Battle.Status != combat.Retreated {
				t.Fatal("unexpected terminal outcome", view.Battle.Status)
			}
			terminal, casts := 0, 0
			for _, event := range view.Events {
				switch event.Kind {
				case "won", "retreated", "defeated", "expired":
					terminal++
				case "player_ability":
					casts++
				}
			}
			var awards, commands, active int
			var respawning bool
			if err := f.s.pool.QueryRow(f.ctx, `SELECT (SELECT count(*) FROM wallet_entries WHERE source='pve'),
 (SELECT count(*) FROM battle_commands),(SELECT count(*) FROM battles WHERE status='active'),respawn_at IS NOT NULL
 FROM npc_spawns WHERE id=$1`, f.in.NPCSpawnID).Scan(&awards, &commands, &active, &respawning); err != nil {
				t.Fatal(err)
			}
			won := view.Battle.Status == combat.Won
			expected := 0
			if won {
				expected = 1
			}
			if terminal != 1 || active != 0 || casts != expected || commands != expected || awards != expected || respawning != won || view.NextActionAt != nil {
				t.Fatalf("partial terminal transition: terminal%d active%d casts%d commands%d awards%d respawn%v", terminal, active, casts, commands, awards, respawning)
			}
			if !won && view.Battle.NPC.HPCurrent != view.Battle.NPC.HP {
				t.Fatal("NPC left damaged after terminal race")
			}
		})
	}
}

func TestCombatExecutionHealingEnergyAndExperienceCap(t *testing.T) {
	t.Run("healing", func(t *testing.T) {
		f := newExecutionFixture(t)
		mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=0,heal=25 WHERE id=$1`, f.abilityID)
		mustExec(t, f.ctx, f.s, `UPDATE characters SET hp_current=95 WHERE id=$1`, f.in.CharacterID)
		result, err := f.cast("heal")
		if err != nil || result.Character.HPCurrent != 100 || result.Battle.NPC.HPCurrent != result.Battle.NPC.HP {
			t.Fatal(result, err)
		}
		event := f.view(t).Events[1]
		if event.Healing != 5 || event.Damage != 0 {
			t.Fatal(event)
		}
	})
	t.Run("energy denial preserves NPC attack", func(t *testing.T) {
		f := newExecutionFixture(t)
		mustExec(t, f.ctx, f.s, `UPDATE characters SET mp_current=0 WHERE id=$1`, f.in.CharacterID)
		f.due(t)
		_, err := f.cast("empty")
		if !errors.Is(err, character.ErrInsufficientEnergy) {
			t.Fatal(err)
		}
		v := f.view(t)
		if len(v.Events) != 2 || v.Events[1].Kind != "npc_ability" {
			t.Fatal(v.Events)
		}
		var ready bool
		if err = f.s.pool.QueryRow(f.ctx, `SELECT global_cooldown_ready_at IS NULL FROM characters WHERE id=$1`, f.in.CharacterID).Scan(&ready); err != nil || !ready {
			t.Fatal("failed cast started GCD", ready, err)
		}
	})
	t.Run("experience cap", func(t *testing.T) {
		f := newExecutionFixture(t)
		mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=10000 WHERE id=$1`, f.abilityID)
		mustExec(t, f.ctx, f.s, `UPDATE characters SET level=(SELECT max(level) FROM levels),experience=(SELECT max(experience_required) FROM levels) WHERE id=$1`, f.in.CharacterID)
		if _, err := f.cast("capped"); err != nil {
			t.Fatal(err)
		}
		event := f.view(t).Events[2]
		if event.Kind != "won" || event.Experience != 0 || event.Credits <= 0 {
			t.Fatal(event)
		}
	})
}

func TestCombatExecutionQueuedUseRechecksSessionAfterAbilityLock(t *testing.T) {
	f := newExecutionFixture(t)
	tx, err := f.s.pool.Begin(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(f.ctx)
	var pid int
	if err = tx.QueryRow(f.ctx, `SELECT pg_backend_pid() FROM character_abilities WHERE character_id=$1 AND ability_id=$2 FOR UPDATE`, f.in.CharacterID, f.abilityID).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := f.cast("queued"); done <- err }()
	awaitResetUserLock(t, f.ctx, f.s, pid, done)
	if _, err = tx.Exec(f.ctx, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, f.actor.SessionHash); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(f.ctx); err != nil {
		t.Fatal(err)
	}
	if err = <-done; !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatal(err)
	}
	var count int
	if err = f.s.pool.QueryRow(f.ctx, `SELECT count(*) FROM battle_events WHERE battle_id=$1`, f.battle.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestCombatExecutionRestartCatchesUpAndPaginatesHistory(t *testing.T) {
	for _, defense := range []int{0, math.MaxInt32} {
		t.Run(map[bool]string{true: "expired", false: "defeated"}[defense > 0], func(t *testing.T) {
			s, ctx, actor, _, in := combatFixture(t)
			var level int
			var rank string
			if err := s.pool.QueryRow(ctx, `SELECT n.level,n.rank FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id WHERE s.id=$1`, in.NPCSpawnID).Scan(&level, &rank); err != nil {
				t.Fatal(err)
			}
			rules, err := combat.RulesFor(level, rank)
			if err != nil {
				t.Fatal(err)
			}
			var created time.Time
			if err = s.pool.QueryRow(ctx, `SELECT clock_timestamp()-interval '1 hour'`).Scan(&created); err != nil {
				t.Fatal(err)
			}
			var id int64
			err = s.pool.QueryRow(ctx, `INSERT INTO battles(character_id,npc_spawn_id,location_id,start_key,npc_template_id,npc_name,npc_level,npc_rank,npc_hp,npc_hp_current,created_at)
 SELECT $1,s.id,n.location_id,'historical',t.id,t.name,n.level,n.rank,$3,$3,$4
 FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id JOIN npc_templates t ON t.id=n.npc_template_id WHERE s.id=$2 RETURNING id`, in.CharacterID, in.NPCSpawnID, rules.HP, created).Scan(&id)
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(rules)
			mustExec(t, ctx, s, `INSERT INTO battle_executions(battle_id,rules,next_action_at,expires_at,created_at) VALUES($1,$2,$3,$4,$5)`, id, encoded, created.Add(time.Duration(rules.Pattern[0].DelayMS)*time.Millisecond), created.Add(10*time.Minute), created)
			mustExec(t, ctx, s, `UPDATE characters SET defense=$2,hp_current=100,hp_updated_at=$3,hp_regen_remainder=0,mp_current=40,mp_updated_at=$3,mp_regen_remainder=0,resources_revision=resources_revision+1 WHERE id=$1`, in.CharacterID, defense, created)
			// A fresh repository has no in-memory knowledge of the old encounter.
			r := db.NewCombatRepository(s.pool)
			worked, err := r.ProcessDue(ctx)
			if err != nil || !worked {
				t.Fatal(worked, err)
			}
			v, err := r.View(ctx, actor, id, in.CharacterID, 0)
			if err != nil {
				t.Fatal(err)
			}
			want := combat.Defeated
			if defense > 0 {
				want = combat.Expired
			}
			if v.Battle.Status != want || v.NextActionAt != nil || v.Character.HPCurrent != 100 || v.Character.MPCurrent != 100 {
				t.Fatal("offline recovery", v.Battle, v.Character.HPCurrent, v.Character.MPCurrent)
			}
			if want == combat.Expired && !v.Battle.EndedAt.Equal(created.Add(10*time.Minute)) {
				t.Fatal("expiry drifted", v.Battle.EndedAt)
			}
			count := 0
			last := created
			for {
				for _, event := range v.Events {
					count++
					if event.Sequence != count || event.OccurredAt.Before(last) {
						t.Fatal("out of order", event)
					}
					last = event.OccurredAt
				}
				if v.NextAfter == nil {
					break
				}
				v, err = r.View(ctx, actor, id, in.CharacterID, *v.NextAfter)
				if err != nil {
					t.Fatal(err)
				}
			}
			if count > 601 || count < 2 {
				t.Fatal("unbounded/lost history", count)
			}
			if defense > 0 && count < 51 {
				t.Fatal("pagination not exercised", count)
			}
			if worked, err = r.ProcessDue(ctx); err != nil || worked {
				t.Fatal("finished encounter reprocessed", worked, err)
			}
		})
	}
}
