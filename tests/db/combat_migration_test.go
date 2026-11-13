package db_test

import (
	"strings"
	"testing"
)

const combatMigration = "20261108000300_battle_lifecycle"

func TestCombatMigrationPreservesWorldAndCharacters(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, combatMigration)
	snapshot := func() string {
		var value string
		if err := s.pool.QueryRow(ctx, `SELECT jsonb_build_array(
		(SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM characters c),
		(SELECT jsonb_agg(to_jsonb(s)-'respawn_at' ORDER BY id) FROM npc_spawns s))::text`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, combatMigration+"."+direction+".sql")
		if snapshot() != before {
			t.Fatalf("%s altered existing state", direction)
		}
	}
	_, err := s.pool.Exec(ctx, `UPDATE npc_spawns SET respawn_at='infinity'`)
	requireConstraint(t, err, "23514", "npc_spawns_respawn_at_check")
}

func TestCombatMigrationRefusesToDiscardBattles(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	first := startBattle(t, ctx, s, actor, in)
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, migrationSQL(t, combatMigration+".down.sql"))
	if err == nil || !strings.Contains(err.Error(), "cannot roll back nonempty battles") {
		t.Fatalf("unsafe down: %v", err)
	}
	if _, err = conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if repeat := startBattle(t, ctx, s, actor, in); !repeat.Replayed || repeat.Battle.ID != first.Battle.ID {
		t.Fatal("failed down lost battle")
	}
}

func TestCombatDatabaseGuardsExclusivityAndFinishedState(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	b := startBattle(t, ctx, s, actor, in).Battle
	insert := `INSERT INTO battles(character_id,npc_spawn_id,location_id,start_key,npc_template_id,npc_name,npc_level,npc_rank,npc_hp,npc_hp_current)
 SELECT $2,$3,location_id,$4,npc_template_id,npc_name,npc_level,npc_rank,npc_hp,npc_hp FROM battles WHERE id=$1`
	_, err := s.pool.Exec(ctx, insert, b.ID, 2, in.NPCSpawnID+1, "same-character")
	requireConstraint(t, err, "23505", "battles_active_character_idx")
	_, err = s.pool.Exec(ctx, insert, b.ID, 3, in.NPCSpawnID, "same-npc")
	requireConstraint(t, err, "23505", "battles_active_spawn_idx")
	_, err = s.pool.Exec(ctx, insert, b.ID, 3, in.NPCSpawnID+1, "bad\n")
	requireConstraint(t, err, "23514", "battles_start_key_check")
	_, err = s.pool.Exec(ctx, `UPDATE battles SET npc_hp_current=npc_hp+1 WHERE id=$1`, b.ID)
	requireConstraint(t, err, "23514", "battles_npc_hp_current_check")
	_, err = s.pool.Exec(ctx, `UPDATE battles SET status='retreated' WHERE id=$1`, b.ID)
	requireConstraint(t, err, "23514", "battles_status_time_check")
	_, err = s.pool.Exec(ctx, `UPDATE battles SET character_id=3 WHERE id=$1`, b.ID)
	requireConstraint(t, err, "23514", "")
	mustExec(t, ctx, s, `UPDATE battles SET status='retreated',ended_at=clock_timestamp() WHERE id=$1`, b.ID)
	_, err = s.pool.Exec(ctx, `UPDATE battles SET status='active',ended_at=NULL WHERE id=$1`, b.ID)
	requireConstraint(t, err, "23514", "")
}
