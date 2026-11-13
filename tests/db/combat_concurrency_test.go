package db_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/world"
)

func TestCombatConcurrentStartReservesNPCOnce(t *testing.T) {
	s, ctx, first, second, in := combatFixture(t)
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, actor := range []character.Actor{first, second} {
		input := in
		input.CharacterID = i + 2
		go func() {
			<-start
			_, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
				return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, input, combat.NPCStats{HP: 200})
			})
			results <- err
		}()
	}
	close(start)
	success, unavailable := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			success++
		} else if errors.Is(err, combat.ErrNPCUnavailable) {
			unavailable++
		} else {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM battles WHERE status='active'`).Scan(&count); err != nil || count != 1 || success != 1 || unavailable != 1 {
		t.Fatalf("active=%d success=%d rejected=%d %v", count, success, unavailable, err)
	}
}

func TestCombatStartRechecksSessionAfterWaitingForNPC(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	blocker, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback(ctx)
	var pid int
	if err = blocker.QueryRow(ctx, `SELECT pg_backend_pid() FROM npc_spawns WHERE id=$1 FOR UPDATE`, in.NPCSpawnID).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
			return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
		})
		done <- err
	}()
	awaitResetUserLock(t, ctx, s, pid, done)
	if _, err = blocker.Exec(ctx, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE token=$1`, actor.SessionHash); err != nil {
		t.Fatal(err)
	}
	if err = blocker.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err = <-done; !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("expired queued start: %v", err)
	}
}

func TestCombatQueuedStartRechecksAccount(t *testing.T) {
	for _, change := range []string{"session", "selection", "ban"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, actor, _, in := combatFixture(t)
			blocker, pid := lockResetUser(t, ctx, s)
			done := make(chan error, 1)
			go func() {
				_, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
					return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
				})
				done <- err
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			query, want := `DELETE FROM sessions WHERE user_id=1`, character.ErrUnauthenticated
			if change == "selection" {
				query, want = `UPDATE users SET character_id=1 WHERE id=1`, character.ErrSelectionChanged
			}
			if change == "ban" {
				query = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`
			}
			if _, err := blocker.Exec(ctx, query); err != nil {
				t.Fatal(err)
			}
			if err := blocker.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, want) {
				t.Fatalf("queued %s: %v", change, err)
			}
		})
	}
}

func TestCombatStartAndMovementCannotBothSucceed(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	target := locationAt(t, ctx, s, "coruscant", 2)
	start := make(chan struct{})
	battleResult, moveResult := make(chan error, 1), make(chan error, 1)
	go func() {
		<-start
		_, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
			return db.NewCombatRepository(s.pool).StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
		})
		battleResult <- err
	}()
	go func() {
		<-start
		_, err := db.NewWorldRepository(s.pool).Move(ctx, actor, world.MoveInput{CharacterID: in.CharacterID, LocationID: target, FromLocationID: in.LocationID})
		moveResult <- err
	}()
	close(start)
	battleErr, moveErr := <-battleResult, <-moveResult
	if !(battleErr == nil && errors.Is(moveErr, combat.ErrInBattle) || moveErr == nil && errors.Is(battleErr, world.ErrPositionChanged)) {
		t.Fatalf("battle=%v move=%v", battleErr, moveErr)
	}
}

func TestCombatRetreatAndDefeatProduceOneTerminalTransition(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	b := startBattle(t, ctx, s, actor, in).Battle
	r := db.NewCombatRepository(s.pool)
	start := make(chan struct{})
	type result struct {
		change *combat.Change
		err    error
	}
	done := make(chan result, 2)
	for _, defeat := range []bool{false, true} {
		go func() {
			<-start
			change, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
				if defeat {
					return r.DefeatInTx(ctx, tx, 2, b.ID)
				}
				return r.RetreatInTx(ctx, tx, actor, 2, b.ID)
			})
			done <- result{change, err}
		}()
	}
	close(start)
	first, second := <-done, <-done
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent endings: %v / %v", first.err, second.err)
	}
	if first.change.Battle.Status != second.change.Battle.Status || first.change.Replayed == second.change.Replayed {
		t.Fatalf("endings: %+v / %+v", first.change, second.change)
	}
	var hp int
	if err := s.pool.QueryRow(ctx, `SELECT hp_current FROM characters WHERE id=2`).Scan(&hp); err != nil {
		t.Fatal(err)
	}
	if first.change.Battle.Status == combat.Defeated && hp != 0 || first.change.Battle.Status == combat.Retreated && hp != 100 {
		t.Fatalf("status=%s hp=%d", first.change.Battle.Status, hp)
	}
}
