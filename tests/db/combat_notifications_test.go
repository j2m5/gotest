package db_test

import (
	"errors"
	"math"
	"testing"

	"gotest/internal/ability"
	"gotest/internal/combat"
	"gotest/internal/db"
)

func TestCombatNotifiesOnlyCommittedChanges(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	mustExec(t, ctx, s, `UPDATE characters SET hp=1000,hp_current=1000 WHERE id=$1`, in.CharacterID)
	mustExec(t, ctx, s, `UPDATE abilities SET cooldown_ms=3600000`)
	mustExec(t, ctx, s, `INSERT INTO character_abilities(character_id,ability_id) SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id ON CONFLICT DO NOTHING`)
	calls := 0
	r := db.NewCombatRepository(s.pool).WithNotifications(func(id int) {
		calls++
		var events int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM battle_events e JOIN battles b ON b.id=e.battle_id WHERE b.character_id=$1`, id).Scan(&events); err != nil || id != in.CharacterID || events < calls {
			t.Errorf("notification before commit or wrong recipient: %d %d %v", id, events, err)
		}
	})
	result, err := r.Start(ctx, actor, in)
	if err != nil || calls != 1 {
		t.Fatal(result, err, calls)
	}
	var abilityID int
	if err = s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=$1 ORDER BY ability_id LIMIT 1`, in.CharacterID).Scan(&abilityID); err != nil {
		t.Fatal(err)
	}
	use := combat.UseInput{CharacterID: in.CharacterID, AbilityID: abilityID, Key: "notify:cast"}
	if _, err = r.Use(ctx, actor, result.Battle.ID, use); err != nil || calls != 2 {
		t.Fatal(err, calls)
	}
	if _, err = r.Use(ctx, actor, result.Battle.ID, use); err != nil || calls != 2 {
		t.Fatal("replay notified", err, calls)
	}
	replay, err := r.Start(ctx, actor, in)
	if err != nil || calls != 2 || !replay.Replayed || replay.EventSequence != 1 {
		t.Fatalf("start acknowledgment: %+v %v calls%d", replay, err, calls)
	}
	mustExec(t, ctx, s, `UPDATE battle_executions SET next_action_at=clock_timestamp() WHERE battle_id=$1`, result.Battle.ID)
	use.Key = "notify:denied"
	if _, err = r.Use(ctx, actor, result.Battle.ID, use); !errors.Is(err, ability.ErrCooldown) || calls != 3 {
		t.Fatal("due attack not notified on rejection", err, calls)
	}
	mustExec(t, ctx, s, `UPDATE battle_executions SET next_action_at=clock_timestamp() WHERE battle_id=$1`, result.Battle.ID)
	if worked, err := r.ProcessDue(ctx); err != nil || !worked || calls != 4 {
		t.Fatal(worked, err, calls)
	}
	retreated, err := r.Retreat(ctx, actor, result.Battle.ID, in.CharacterID)
	if err != nil || calls != 5 {
		t.Fatalf("retreat: %+v battle=%+v err=%v calls=%d", retreated, retreated.Battle, err, calls)
	}
	if _, err = r.Retreat(ctx, actor, result.Battle.ID, in.CharacterID); err != nil || calls != 5 {
		t.Fatal("repeat retreat notified", err, calls)
	}
}
func TestCombatRollbackDoesNotNotify(t *testing.T) {
	f := newExecutionFixture(t)
	calls := 0
	f.r = f.r.WithNotifications(func(int) { calls++ })
	mustExec(t, f.ctx, f.s, `UPDATE abilities SET damage=100000 WHERE id=$1`, f.abilityID)
	mustExec(t, f.ctx, f.s, `UPDATE wallets SET balance=$1 WHERE character_id=$2 AND currency='credits'`, int64(math.MaxInt64), f.in.CharacterID)
	if _, err := f.cast("overflow:notify"); err == nil {
		t.Fatal("expected rollback")
	}
	if calls != 0 {
		t.Fatalf("rollback published %d notifications", calls)
	}
}
