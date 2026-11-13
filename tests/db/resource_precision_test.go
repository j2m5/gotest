package db_test

import (
	"errors"
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/db"
)

const precisionMigration = "20261108000100_resource_precision"

func TestResourcePrecisionMigrationPreservesState(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, precisionMigration)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=37,mp_current=14 WHERE id=1`)
	snapshot := func() string {
		var result string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(c)-'hp_regen_remainder'-'mp_regen_remainder'-'resources_revision' ORDER BY id)::text FROM characters c`).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	before := snapshot()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, precisionMigration+"."+direction+".sql")
		if snapshot() != before {
			t.Fatalf("%s modified existing state", direction)
		}
	}
	_, err := s.pool.Exec(ctx, `UPDATE characters SET hp_regen_remainder=300000 WHERE id=1`)
	requireConstraint(t, err, "23514", "characters_hp_remainder_check")
	_, err = s.pool.Exec(ctx, `UPDATE characters SET mp_regen_remainder=-1 WHERE id=1`)
	requireConstraint(t, err, "23514", "characters_mp_remainder_check")
	_, err = s.pool.Exec(ctx, `UPDATE characters SET hp_regen_remainder=1 WHERE id=2`)
	requireConstraint(t, err, "23514", "characters_hp_remainder_check")
}

func TestResourcePrecisionManagedWritesKeepSameTimeCarry(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=50,mp_current=20 WHERE id=1;
	 UPDATE characters SET hp_regen_remainder=250000,mp_regen_remainder=50000,
	 hp_updated_at=clock_timestamp()+interval '1 hour',mp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=1`)
	actor := repositoryActor(t, ctx, s, 1, "precision")
	before := characterStatsSnapshot(t, ctx, s)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, want := range []int{49, 48} {
		c, err := db.ChangeCharacterResourcesInTx(ctx, tx, 1, character.ResourceChange{HP: -1, MP: -1})
		if err != nil || c.HPCurrent != want || c.Regeneration.HPRemainder != 250000 || c.Regeneration.MPRemainder != 50000 {
			t.Fatalf("managed carry: %+v %v", c, err)
		}
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("rollback did not restore resource state")
	}
	tx, err = s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = db.ChangeCharacterResourcesInTx(ctx, tx, 1, character.ResourceChange{HP: -1, MP: -1}); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	roster, err := db.NewCharacterRepository(s.pool).List(ctx, actor)
	if err != nil || roster.Characters[0].HPCurrent != 49 || roster.Characters[0].Regeneration.HPRemainder != 250000 {
		t.Fatalf("persisted read: %+v %v", roster, err)
	}
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=48 WHERE id=1`)
	var hpCarry, mpCarry int64
	var hpAt, mpAt time.Time
	if err = s.pool.QueryRow(ctx, `SELECT hp_regen_remainder,mp_regen_remainder,hp_updated_at,mp_updated_at FROM characters WHERE id=1`).Scan(&hpCarry, &mpCarry, &hpAt, &mpAt); err != nil {
		t.Fatal(err)
	}
	if hpCarry != 0 || mpCarry != 50000 || !hpAt.Before(mpAt) {
		t.Fatal("raw HP replacement must reset only HP")
	}
}

func TestResourcePrecisionErrorsDoNotWrite(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE characters SET mp_current=5 WHERE id=1;
	 UPDATE characters SET mp_updated_at=clock_timestamp()+interval '1 hour' WHERE id=1`)
	before := characterStatsSnapshot(t, ctx, s)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = db.ChangeCharacterResourcesInTx(ctx, tx, 1, character.ResourceChange{HP: -10, MP: -6})
	if !errors.Is(err, character.ErrInsufficientEnergy) {
		t.Fatalf("overspend: %v", err)
	}
	_, err = db.ChangeCharacterResourcesInTx(ctx, tx, 999999, character.ResourceChange{HP: -1})
	if !errors.Is(err, character.ErrNotFound) {
		t.Fatalf("missing character: %v", err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("failed operations changed state")
	}
}
