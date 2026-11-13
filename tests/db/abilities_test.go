package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gotest/internal/ability"
	"gotest/internal/character"
	"gotest/internal/db"
)

func abilityFixture(t *testing.T) (*testStorage, context.Context, character.Actor, int) {
	t.Helper()
	s, ctx := characterTestStorage(t, true)
	actor := repositoryActor(t, ctx, s, 1, "abilities")
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1;
	 INSERT INTO character_abilities(character_id,ability_id)
	 SELECT c.id,a.ability_id FROM characters c JOIN class_abilities a ON a.class_id=c.class_id ON CONFLICT DO NOTHING`)
	var id int
	if err := s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=1`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return s, ctx, actor, id
}

func TestAbilityLoadoutChecksSelectionAndDoesNotWrite(t *testing.T) {
	s, ctx, actor, id := abilityFixture(t)
	r := db.NewAbilityRepository(s.pool)
	before := characterStatsSnapshot(t, ctx, s)
	loadout, err := r.List(ctx, actor, 1)
	if err != nil || len(loadout.Abilities) != 1 || loadout.Abilities[0].ID != id || loadout.Abilities[0].Availability != ability.Ready || loadout.Character.ID != 1 || !loadout.ObservedAt.Equal(loadout.Character.Regeneration.ObservedAt) {
		t.Fatalf("loadout=%+v err=%v", loadout, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("reading loadout wrote character state")
	}
	if _, err = r.List(ctx, actor, 2); !errors.Is(err, character.ErrSelectionChanged) {
		t.Fatalf("stale character: %v", err)
	}
	if _, err = r.List(ctx, actor, 3); !errors.Is(err, character.ErrSelectionChanged) {
		t.Fatalf("other account: %v", err)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=NULL WHERE id=1`)
	if _, err = r.List(ctx, actor, 1); !errors.Is(err, character.ErrSelectionRequired) {
		t.Fatalf("no selection: %v", err)
	}
}

func TestAbilityPreparationPersistsCooldownsAndComposesRollback(t *testing.T) {
	s, ctx, actor, id := abilityFixture(t)
	r := db.NewAbilityRepository(s.pool)
	mustExec(t, ctx, s, `UPDATE abilities SET cooldown_ms=60000 WHERE code='sith_warrior_basic_attack'`)
	before := characterStatsSnapshot(t, ctx, s)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	use, err := r.PrepareUseInTx(ctx, tx, 1, id)
	if err != nil || use.EnergyRemaining != 90 || use.ReadyAt.Sub(use.UsedAt) != time.Minute || use.GlobalReadyAt.Sub(use.UsedAt) != 1500*time.Millisecond {
		t.Fatalf("use=%+v err=%v", use, err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("aborted combat charged energy or started cooldown")
	}
	tx, err = s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	use, err = r.PrepareUseInTx(ctx, tx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	characters := db.NewCharacterRepository(s.pool)
	if _, err = characters.SetActive(ctx, actor, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = characters.SetActive(ctx, actor, new(1)); err != nil {
		t.Fatal(err)
	}
	loadout, err := db.NewAbilityRepository(s.pool).List(ctx, actor, 1)
	if err != nil || loadout.GlobalReadyAt == nil || !loadout.GlobalReadyAt.Equal(use.GlobalReadyAt) || len(loadout.Abilities) != 1 || loadout.Abilities[0].ReadyAt == nil || !loadout.Abilities[0].ReadyAt.Equal(use.ReadyAt) || loadout.Abilities[0].Availability != ability.Cooldown {
		t.Fatalf("persisted cooldown: %+v %v", loadout, err)
	}
}

func TestAbilityLoadoutFiltersClassAndRechecksAuthentication(t *testing.T) {
	s, ctx, actor, id := abilityFixture(t)
	r := db.NewAbilityRepository(s.pool)
	mustExec(t, ctx, s, `INSERT INTO character_abilities(character_id,ability_id) SELECT 1,id FROM abilities WHERE code='trooper_basic_attack';
	 INSERT INTO abilities(code,name,damage,mp_cost) VALUES('unlearned','Unlearned',10,0);
	 INSERT INTO class_abilities SELECT c.id,a.id FROM classes c CROSS JOIN abilities a WHERE c.code='sith_warrior' AND a.code='unlearned'`)
	loadout, err := r.List(ctx, actor, 1)
	if err != nil || len(loadout.Abilities) != 1 || loadout.Abilities[0].ID != id {
		t.Fatalf("unlearned or incompatible skill leaked: %+v %v", loadout, err)
	}
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
	if _, err = r.List(ctx, actor, 1); !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("expired session: %v", err)
	}
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()+interval '1 hour' WHERE user_id=1;
	 INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`)
	if _, err = r.List(ctx, actor, 1); !errors.Is(err, character.ErrUnauthenticated) {
		t.Fatalf("account ban: %v", err)
	}
}

func TestAbilityPreparationKeepsFractionalEnergy(t *testing.T) {
	s, ctx, _, id := abilityFixture(t)
	mustExec(t, ctx, s, `UPDATE characters SET mp_current=10 WHERE id=1;
	 UPDATE characters SET mp_regen_remainder=50000,mp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=1`)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	use, err := db.NewAbilityRepository(s.pool).PrepareUseInTx(ctx, tx, 1, id)
	if err != nil || use.EnergyRemaining != 0 {
		t.Fatalf("use=%+v err=%v", use, err)
	}
	var energy, remainder int
	if err = tx.QueryRow(ctx, `SELECT mp_current,mp_regen_remainder FROM characters WHERE id=1`).Scan(&energy, &remainder); err != nil || energy != 0 || remainder != 50000 {
		t.Fatalf("energy=%d remainder=%d err=%v", energy, remainder, err)
	}
}

func TestAbilityGlobalCooldownProtectsDifferentAbilities(t *testing.T) {
	s, ctx, _, first := abilityFixture(t)
	mustExec(t, ctx, s, `INSERT INTO abilities(code,name,damage,mp_cost,cooldown_ms) VALUES ('ability_test_second','Second',30,20,0);
	 INSERT INTO class_abilities SELECT c.id,a.id FROM classes c CROSS JOIN abilities a WHERE c.code='sith_warrior' AND a.code='ability_test_second';
	 INSERT INTO character_abilities SELECT 1,id FROM abilities WHERE code='ability_test_second'`)
	var second int
	if err := s.pool.QueryRow(ctx, `SELECT id FROM abilities WHERE code='ability_test_second'`).Scan(&second); err != nil {
		t.Fatal(err)
	}
	r := db.NewAbilityRepository(s.pool)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = r.PrepareUseInTx(ctx, tx, 1, first); err != nil {
		t.Fatal(err)
	}
	if _, err = r.PrepareUseInTx(ctx, tx, 1, second); !errors.Is(err, ability.ErrGlobalCooldown) {
		t.Fatalf("second skill bypassed GCD: %v", err)
	}
	var energy int
	if err = tx.QueryRow(ctx, `SELECT mp_current FROM characters WHERE id=1`).Scan(&energy); err != nil || energy != 90 {
		t.Fatalf("energy=%d err=%v", energy, err)
	}
}

func TestAbilityInvalidUsesNeverMutateState(t *testing.T) {
	s, ctx, _, id := abilityFixture(t)
	r := db.NewAbilityRepository(s.pool)
	var other int
	if err := s.pool.QueryRow(ctx, `SELECT ability_id FROM character_abilities WHERE character_id=3`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, setup string
		char, skill int
		want        error
	}{
		{"not learned", "", 1, other, ability.ErrUnavailable},
		{"missing character", "", 999999, id, character.ErrNotFound},
		{"insufficient energy", `UPDATE characters SET mp_current=5 WHERE id=1; UPDATE characters SET mp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=1`, 1, id, character.ErrInsufficientEnergy},
		{"defeated", `UPDATE characters SET hp_current=0 WHERE id=1; UPDATE characters SET hp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=1`, 1, id, ability.ErrDefeated},
		{"wrong class even if learned", `INSERT INTO character_abilities(character_id,ability_id) SELECT 1,id FROM abilities WHERE code='trooper_basic_attack'`, 1, other, ability.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != "" {
				mustExec(t, ctx, s, tc.setup)
			}
			before := characterStatsSnapshot(t, ctx, s)
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if _, err = r.PrepareUseInTx(ctx, tx, tc.char, tc.skill); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if characterStatsSnapshot(t, ctx, s) != before {
				t.Fatal("invalid use wrote state")
			}
		})
	}
}

func TestAbilityConcurrentUseCommitsOnce(t *testing.T) {
	s, ctx, _, id := abilityFixture(t)
	mustExec(t, ctx, s, `UPDATE abilities SET cooldown_ms=60000 WHERE code='sith_warrior_basic_attack'`)
	r := db.NewAbilityRepository(s.pool)
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				results <- err
				return
			}
			defer tx.Rollback(ctx)
			_, err = r.PrepareUseInTx(ctx, tx, 1, id)
			if err == nil {
				err = tx.Commit(ctx)
			}
			results <- err
		}()
	}
	close(start)
	success, rejected := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			success++
		} else if errors.Is(err, ability.ErrCooldown) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	var energy int
	if err := s.pool.QueryRow(ctx, `SELECT mp_current FROM characters WHERE id=1`).Scan(&energy); err != nil || energy != 90 || success != 1 || rejected != 1 {
		t.Fatalf("success=%d rejected=%d energy=%d err=%v", success, rejected, energy, err)
	}
}
