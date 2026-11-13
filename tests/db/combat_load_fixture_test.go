package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/ability"
	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
)

type loadActor struct {
	userID, characterID, abilityID int
	battleID                       int64
	token                          string
}

func combatLoadFixture(t testing.TB, s *testStorage, ctx context.Context, count int) []loadActor {
	t.Helper()
	var locationID, placementID, templateID, slot int
	var npcName string
	err := s.pool.QueryRow(ctx, `SELECT n.location_id,n.id,n.npc_template_id,t.name,
 (SELECT max(slot) FROM npc_spawns WHERE location_npc_id=n.id)
 FROM location_npcs n JOIN npc_templates t ON t.id=n.npc_template_id
 JOIN locations l ON l.id=n.location_id JOIN planets p ON p.id=l.planet_id
 WHERE p.code='coruscant' AND n.level=1 AND n.rank='normal' ORDER BY n.id LIMIT 1`).Scan(&locationID, &placementID, &templateID, &npcName, &slot)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at)
 SELECT 'load'||i||'@example.test','load'||i,'unused',clock_timestamp() FROM generate_series(1,$1) i`, count)
	mustExec(t, ctx, s, `INSERT INTO characters(user_id,faction_id,class_id,name,location_id,hp,hp_current,power,mastery)
 SELECT u.id,c.faction_id,c.id,'Load '||u.id,$1,100000,100000,0,0 FROM users u
 CROSS JOIN classes c WHERE u.login LIKE 'load%' AND c.code='jedi_consular' ORDER BY u.id`, locationID)
	mustExec(t, ctx, s, `UPDATE users u SET character_id=c.id FROM characters c WHERE c.user_id=u.id AND u.login LIKE 'load%';
 INSERT INTO character_abilities(character_id,ability_id)
 SELECT c.id,a.ability_id FROM characters c JOIN users u ON u.id=c.user_id
 JOIN class_abilities a ON a.class_id=c.class_id WHERE u.login LIKE 'load%'`)
	mustExec(t, ctx, s, `INSERT INTO npc_spawns(location_npc_id,slot) SELECT $1,$2+i FROM generate_series(1,$3) i`, placementID, slot, count)
	rows, err := s.pool.Query(ctx, `SELECT u.id,c.id,a.ability_id FROM users u
 JOIN characters c ON c.id=u.character_id JOIN character_abilities a ON a.character_id=c.id
 WHERE u.login LIKE 'load%' ORDER BY u.id`)
	if err != nil {
		t.Fatal(err)
	}
	actors, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (loadActor, error) {
		var actor loadActor
		err := row.Scan(&actor.userID, &actor.characterID, &actor.abilityID)
		actor.battleID = int64(actor.characterID)
		actor.token = fmt.Sprintf("%064x", actor.userID)
		return actor, err
	})
	if err != nil || len(actors) != count {
		t.Fatalf("load actors: got %d, want %d: %v", len(actors), count, err)
	}
	rows, err = s.pool.Query(ctx, `SELECT id FROM npc_spawns WHERE location_npc_id=$1 AND slot>$2 ORDER BY slot`, placementID, slot)
	if err != nil {
		t.Fatal(err)
	}
	spawns, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil || len(spawns) != count {
		t.Fatal("load spawn fixture", err)
	}
	var created time.Time
	if err = s.pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&created); err != nil {
		t.Fatal(err)
	}
	rules, err := combat.RulesFor(1, "normal")
	if err != nil {
		t.Fatal(err)
	}
	rules.HP = 100000 // Sustain the observation window; keep attack timing/damage.
	encoded, err := json.Marshal(rules)
	if err != nil {
		t.Fatal(err)
	}
	var sessions, battles, executions, events [][]any
	for i, a := range actors {
		sessions = append(sessions, []any{a.userID, auth.TokenDigest(a.token), created.Add(time.Hour)})
		battles = append(battles, []any{a.battleID, a.characterID, spawns[i], locationID, "load:start", templateID, npcName, 1, "normal", rules.HP, rules.HP, created})
		executions = append(executions, []any{a.battleID, encoded, created.Add(5 * time.Second), created.Add(10 * time.Minute), 1, created})
		payload, err := json.Marshal(combat.Event{Sequence: 1, Kind: "started", Actor: "player", CharacterHP: 100000, CharacterMP: 100, NPCHP: rules.HP, OccurredAt: created})
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, []any{a.battleID, 1, created, payload})
	}
	copyRows := func(table string, columns []string, data [][]any) {
		t.Helper()
		if _, err := s.pool.CopyFrom(ctx, pgx.Identifier{table}, columns, pgx.CopyFromRows(data)); err != nil {
			t.Fatal(table, err)
		}
	}
	copyRows("sessions", []string{"user_id", "token", "expires_at"}, sessions)
	copyRows("battles", []string{"id", "character_id", "npc_spawn_id", "location_id", "start_key", "npc_template_id", "npc_name", "npc_level", "npc_rank", "npc_hp", "npc_hp_current", "created_at"}, battles)
	copyRows("battle_executions", []string{"battle_id", "rules", "next_action_at", "expires_at", "event_sequence", "created_at"}, executions)
	copyRows("battle_events", []string{"battle_id", "sequence", "occurred_at", "payload"}, events)
	// Only the fixture tables, never unqualified ANALYZE of the whole database.
	for _, table := range []string{"users", "sessions", "characters", "character_abilities", "battles", "battle_executions", "battle_events", "npc_spawns"} {
		mustExec(t, ctx, s, "ANALYZE "+pgx.Identifier{s.pool.Config().ConnConfig.RuntimeParams["search_path"], table}.Sanitize())
	}
	return actors
}

func combatLoadRouter(t testing.TB, s *testStorage, repository *db.CombatRepository) http.Handler {
	t.Helper()
	a := auth.NewService(s.AuthRepository, auth.Options{})
	h, err := routes.New(handlers.NewHandler(a, false, a),
		handlers.NewCharacterHandler(character.NewService(db.NewCharacterRepository(s.pool))), middleware.NewMiddleware(a), routes.Options{
			StaticDir: t.TempDir(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			Abilities: handlers.NewAbilityHandler(ability.NewService(db.NewAbilityRepository(s.pool))),
			Combat:    handlers.NewCombatHandler(combat.NewService(repository)),
		})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
