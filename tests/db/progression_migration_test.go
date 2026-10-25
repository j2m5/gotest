package db_test

import (
	"strings"
	"testing"

	"gotest/internal/db"
	"gotest/internal/progression"
)

const experienceRewardMigration = "20261025000800_experience_rewards"

func TestExperienceRewardMigrationPreservesExistingData(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, experienceRewardMigration)
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1; UPDATE characters SET level=3,experience=300,hp_current=5 WHERE id=1`)
	before := characterStatsSnapshot(t, ctx, s)
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, experienceRewardMigration+"."+direction+".sql")
		if characterStatsSnapshot(t, ctx, s) != before {
			t.Fatalf("%s changed existing data", direction)
		}
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT to_regclass('character_experience_rewards') IS NOT NULL`).Scan(&exists); err != nil || exists != (direction == "up") {
			t.Fatalf("%s ledger exists = %t, %v", direction, exists, err)
		}
	}
}

func TestExperienceRewardMigrationRefusesToDiscardHistory(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	reward := progression.Reward{CharacterID: 1, Key: "quest:keep-history", Experience: 100}
	if _, err := service.Grant(ctx, reward); err != nil {
		t.Fatal(err)
	}
	before := characterStatsSnapshot(t, ctx, s)
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, migrationSQL(t, experienceRewardMigration+".down.sql"))
	if err == nil || !strings.Contains(err.Error(), "cannot roll back a nonempty reward ledger") {
		t.Fatalf("unsafe down = %v", err)
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("failed down changed character data")
	}
	assertRewardCount(t, ctx, s, 1)
	result, err := service.Grant(ctx, reward)
	if err != nil || result == nil || !result.Replayed {
		t.Fatalf("down lost replay protection = %+v, %v", result, err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 2, Experience: 100})
}

func TestExperienceRewardLedgerConstraintsAndCharacterCascade(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	service := progression.NewService(db.NewProgressionRepository(s.pool))
	for id := 1; id <= 3; id++ {
		if _, err := service.Grant(ctx, progression.Reward{CharacterID: id, Key: "quest:shared", Experience: 100}); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct{ name, assignment, constraint string }{
		{"empty key", `reward_key=''`, "key"},
		{"non ASCII key", `reward_key='quest:é'`, "key"},
		{"space in key", `reward_key='quest:1 '`, "key"},
		{"leading separator", `reward_key=':quest'`, "key"},
		{"trailing newline", `reward_key=E'quest:1\n'`, "key"},
		{"zero requested", `requested_amount=0,applied_amount=0,experience_after=experience_before`, "requested"},
		{"negative applied", `applied_amount=-1`, "applied"},
		{"excess applied", `applied_amount=101,experience_after=101`, "applied"},
		{"negative experience", `experience_before=-1,experience_after=99`, "experience"},
		{"lost experience", `experience_before=200,experience_after=100`, "experience"},
		{"incorrect delta", `experience_after=101`, "experience"},
		{"zero level", `level_before=0`, "levels"},
		{"lost level", `level_before=3,level_after=2`, "levels"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.pool.Exec(ctx, `UPDATE character_experience_rewards SET `+tt.assignment+` WHERE character_id=1`)
			requireConstraint(t, err, "23514", "character_experience_rewards_"+tt.constraint+"_check")
		})
	}
	// Key validation has identical byte rules in Go and SQL, including boundaries.
	for _, key := range []string{"a", strings.Repeat("Z", 128), "Quest:123/attempt_7.1-2"} {
		if _, err := service.Grant(ctx, progression.Reward{CharacterID: 1, Key: key, Experience: 1}); err != nil {
			t.Fatalf("valid key %q: %v", key, err)
		}
	}
	_, err := s.pool.Exec(ctx, `UPDATE character_experience_rewards SET character_id=999999 WHERE character_id=3`)
	requireConstraint(t, err, "23503", "character_experience_rewards_character_id_fkey")
	_, err = s.pool.Exec(ctx, `UPDATE character_experience_rewards SET character_id=2 WHERE character_id=3`)
	requireConstraint(t, err, "23505", "character_experience_rewards_pkey")
	mustExec(t, ctx, s, `DELETE FROM characters WHERE id=1`)
	assertRewardCount(t, ctx, s, 2)
	mustExec(t, ctx, s, `DELETE FROM users WHERE id=1`)
	assertRewardCount(t, ctx, s, 1)
	assertProgressionState(t, ctx, s, 3, progression.State{Level: 2, Experience: 100})
}
