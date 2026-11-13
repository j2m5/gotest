package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/world"
)

func combatFixture(t *testing.T) (*testStorage, context.Context, character.Actor, character.Actor, combat.StartInput) {
	t.Helper()
	s, ctx := characterTestStorage(t, true)
	first := repositoryActor(t, ctx, s, 1, "combat-one")
	second := repositoryActor(t, ctx, s, 2, "combat-two")
	mustExec(t, ctx, s, `UPDATE users SET character_id=CASE id WHEN 1 THEN 2 WHEN 2 THEN 3 END WHERE id IN(1,2)`)
	input := combat.StartInput{CharacterID: 2, LocationID: locationAt(t, ctx, s, "coruscant", 1), Key: "combat:first"}
	if err := s.pool.QueryRow(ctx, `SELECT s.id FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id WHERE n.location_id=$1 ORDER BY s.id LIMIT 1`, input.LocationID).Scan(&input.NPCSpawnID); err != nil {
		t.Fatal(err)
	}
	return s, ctx, first, second, input
}

func startBattle(t *testing.T, ctx context.Context, s *testStorage, actor character.Actor, in combat.StartInput) *combat.Change {
	t.Helper()
	result, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
		return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func combatTransaction(ctx context.Context, s *testStorage, action func(pgx.Tx) (*combat.Change, error)) (*combat.Change, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	result, err := action(tx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func TestCombatStartIsPersistentIdempotentAndReadOnly(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	r := db.NewCombatRepository(s.pool)
	state, err := r.Current(ctx, actor, in.CharacterID)
	if err != nil || state.Battle != nil || state.Character.ID != in.CharacterID {
		t.Fatalf("empty state: %+v %v", state, err)
	}
	first := startBattle(t, ctx, s, actor, in)
	if first.Replayed || first.Battle.Status != combat.Active || first.Battle.NPC.HPCurrent != 200 || first.Battle.EndedAt != nil {
		t.Fatalf("start: %+v", first)
	}
	again := startBattle(t, ctx, s, actor, in)
	if !again.Replayed || again.Battle.ID != first.Battle.ID {
		t.Fatal("start duplicated")
	}
	before := characterStatsSnapshot(t, ctx, s)
	state, err = db.NewCombatRepository(s.pool).Current(ctx, actor, in.CharacterID)
	if err != nil || state.Battle == nil || state.Battle.ID != first.Battle.ID || state.Battle.NPC.Name == "" {
		t.Fatalf("persisted battle: %+v %v", state, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("Current wrote resource or session state")
	}
	changed := in
	changed.NPCSpawnID++
	_, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
		return r.StartInTx(ctx, tx, actor, changed, combat.NPCStats{HP: 200})
	})
	if !errors.Is(err, combat.ErrConflict) {
		t.Fatalf("key reused for another target: %v", err)
	}
	changed = in
	changed.Key = "another"
	_, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
		return r.StartInTx(ctx, tx, actor, changed, combat.NPCStats{HP: 200})
	})
	if !errors.Is(err, combat.ErrInBattle) {
		t.Fatalf("second battle: %v", err)
	}
}

func TestCombatStartRollbackLeavesNPCFree(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	before := characterStatsSnapshot(t, ctx, s)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200}); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM battles`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback: %d %v", count, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("start changed character state")
	}
	if startBattle(t, ctx, s, actor, in).Replayed {
		t.Fatal("aborted start retained its key")
	}
}

func TestCombatStartValidatesSelectionPositionAndAvailability(t *testing.T) {
	for _, name := range []string{"selection", "no selection", "wrong session", "expired", "banned", "position", "foreign NPC", "respawning", "defeated"} {
		t.Run(name, func(t *testing.T) {
			s, ctx, actor, _, in := combatFixture(t)
			want := character.ErrUnauthenticated
			switch name {
			case "selection":
				mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
				want = character.ErrSelectionChanged
			case "no selection":
				mustExec(t, ctx, s, `UPDATE users SET character_id=NULL WHERE id=1`)
				want = character.ErrSelectionRequired
			case "wrong session":
				actor.UserID = 2
			case "expired":
				mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
			case "banned":
				mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`)
			case "position":
				in.LocationID++
				want = world.ErrPositionChanged
			case "foreign NPC":
				if err := s.pool.QueryRow(ctx, `SELECT s.id FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id WHERE n.location_id<>$1 LIMIT 1`, in.LocationID).Scan(&in.NPCSpawnID); err != nil {
					t.Fatal(err)
				}
				want = combat.ErrNPCUnavailable
			case "respawning":
				mustExec(t, ctx, s, `UPDATE npc_spawns SET respawn_at=clock_timestamp()+interval '1 hour' WHERE id=$1`, in.NPCSpawnID)
				want = combat.ErrNPCUnavailable
			case "defeated":
				mustExec(t, ctx, s, `UPDATE characters SET hp_current=0 WHERE id=2; UPDATE characters SET hp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=2`)
				want = combat.ErrDefeated
			}
			_, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
				return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
			})
			if !errors.Is(err, want) {
				t.Fatalf("got %v want %v", err, want)
			}
			var count int
			if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM battles`).Scan(&count); err != nil || count != 0 {
				t.Fatalf("invalid start created battle: %d %v", count, err)
			}
		})
	}
}
