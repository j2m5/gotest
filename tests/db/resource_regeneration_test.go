package db_test

import (
	"testing"
	"time"

	"gotest/internal/character"
	"gotest/internal/db"
)

const resourceMigration = "20261029000200_resource_regeneration"

func TestResourceRegenerationReadsDoNotWriteOrResetProgress(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	repository := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 1, "resource-recovery")
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=0,mp_current=0 WHERE user_id=1;
		UPDATE characters SET hp_updated_at=clock_timestamp()-interval '150 seconds',mp_updated_at=clock_timestamp()-interval '30 seconds' WHERE user_id=1`)
	before := characterStatsSnapshot(t, ctx, s)
	for read := 0; read < 3; read++ {
		roster, err := repository.List(ctx, actor)
		if err != nil {
			t.Fatal(err)
		}
		if roster.ActiveCharacterID != nil || len(roster.Characters) != 2 {
			t.Fatalf("roster=%+v", roster)
		}
		for _, c := range roster.Characters {
			if c.HPCurrent < 50 || c.HPCurrent > 51 || c.MPCurrent < 50 || c.MPCurrent > 55 {
				t.Fatalf("recovery=%+v", c)
			}
			r := c.Regeneration
			if r.HPBaseline != 0 || r.MPBaseline != 0 || r.HPFullRestoreMS != 300000 || r.MPFullRestoreMS != 60000 {
				t.Fatalf("metadata=%+v", r)
			}
			if c.HPCurrent != character.RecoverResource(0, c.HP, r.HPUpdatedAt, r.ObservedAt, character.HealthRecoveryPeriod) {
				t.Fatal("snapshot clock mismatch")
			}
		}
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("reads modified persisted baselines or anchors")
	}
	selected, err := repository.SetActive(ctx, actor, new(1))
	if err != nil || selected.HPCurrent < 50 {
		t.Fatalf("selected=%+v err=%v", selected, err)
	}
	mustExec(t, ctx, s, `UPDATE characters SET hp_updated_at=clock_timestamp()-interval '1 year',mp_updated_at=clock_timestamp()-interval '1 year' WHERE id=1`)
	restarted := db.NewCharacterRepository(s.pool)
	selected, err = restarted.SetActive(ctx, actor, new(1))
	if err != nil || selected.HPCurrent != selected.HP || selected.MPCurrent != selected.MP {
		t.Fatalf("offline recovery=%+v err=%v", selected, err)
	}
}

func TestResourceRegenerationAnchorsTrackOnlyChangedResources(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=20,mp_current=30 WHERE id=1;
		UPDATE characters SET hp_updated_at='2026-01-01Z',mp_updated_at='2026-01-02Z' WHERE id=1`)
	anchors := func() (time.Time, time.Time) {
		var hp, mp time.Time
		if err := s.pool.QueryRow(ctx, `SELECT hp_updated_at,mp_updated_at FROM characters WHERE id=1`).Scan(&hp, &mp); err != nil {
			t.Fatal(err)
		}
		return hp, mp
	}
	hp, mp := anchors()
	mustExec(t, ctx, s, `UPDATE characters SET avatar='images/icons/neutral.png',experience=1 WHERE id=1;
		UPDATE characters SET hp_current=hp_current,mp=mp WHERE id=1`)
	if h, m := anchors(); !h.Equal(hp) || !m.Equal(mp) {
		t.Fatal("unrelated/no-op write reset recovery")
	}
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=19 WHERE id=1`)
	newHP, newMP := anchors()
	if !newHP.After(hp) || !newMP.Equal(mp) {
		t.Fatal("HP change must reset only HP anchor")
	}
	mustExec(t, ctx, s, `UPDATE characters SET mp=200 WHERE id=1`)
	if h, m := anchors(); !h.Equal(newHP) || !m.After(mp) {
		t.Fatal("energy capacity change must reset only energy anchor")
	}
	_, err := s.pool.Exec(ctx, `UPDATE characters SET hp_updated_at='infinity' WHERE id=1`)
	requireConstraint(t, err, "23514", "characters_resource_times_finite_check")
}

func TestResourceRegenerationConcurrentWriteHasCoherentBaseline(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	repository := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 1, "resource-transaction")
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE characters SET hp_current=0 WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	before, err := repository.List(ctx, actor)
	if err != nil || before.Characters[0].HPCurrent != 100 {
		t.Fatalf("uncommitted write leaked: %+v %v", before, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := repository.List(ctx, actor)
	if err != nil || after.Characters[0].Regeneration.HPBaseline != 0 || after.Characters[0].HPCurrent > 1 {
		t.Fatalf("new baseline/anchor not read together: %+v %v", after, err)
	}
}

func TestResourceRegenerationMigrationRoundTrip(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, resourceMigration)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=0,mp=0,mp_current=0 WHERE id=1`)
	rows := func() string {
		var result string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(c)-'hp_updated_at'-'mp_updated_at' ORDER BY id)::text FROM characters c`).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	before := rows()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, resourceMigration+"."+direction+".sql")
		if rows() != before {
			t.Fatalf("%s changed existing data", direction)
		}
		var columns int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='characters' AND column_name IN ('hp_updated_at','mp_updated_at') AND data_type='timestamp with time zone' AND is_nullable='NO'`).Scan(&columns); err != nil {
			t.Fatal(err)
		}
		want := 2
		if direction == "down" {
			want = 0
		}
		if columns != want {
			t.Fatalf("%s columns=%d", direction, columns)
		}
	}
}
