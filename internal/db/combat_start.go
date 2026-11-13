package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/world"
)

// StartInTx reserves a shared NPC in the caller's READ COMMITTED transaction.
// Stats must come from trusted server balance, never HTTP input. The future
// combat executor initializes its pattern/events in this same transaction before
// committing. Roll back on every error. No HTTP start is exposed until it can run.
func (r *CombatRepository) StartInTx(ctx context.Context, tx pgx.Tx, actor character.Actor, in combat.StartInput, stats combat.NPCStats) (*combat.Change, error) {
	return r.startCombatInTx(ctx, tx, actor, in, stats, false)
}

func (r *CombatRepository) startCombatInTx(ctx context.Context, tx pgx.Tx, actor character.Actor, in combat.StartInput, stats combat.NPCStats, deriveRules bool) (*combat.Change, error) {
	if err := combat.ValidateStart(in, stats); err != nil {
		return nil, err
	}
	if err := authenticateGameActor(ctx, tx, actor, true); err != nil {
		return nil, err
	}
	faction, locationID, err := worldCharacter(ctx, tx, actor, in.CharacterID, true)
	if err != nil {
		return nil, err
	}
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	previous, err := scanBattle(tx.QueryRow(ctx, `SELECT `+battleColumns+` FROM battles WHERE character_id=$1 AND start_key=$2`, in.CharacterID, in.Key))
	if err == nil {
		if previous.LocationID != in.LocationID || previous.NPC.SpawnID != in.NPCSpawnID {
			return nil, combat.ErrConflict
		}
		return &combat.Change{Battle: previous, Replayed: true}, nil
	}
	if !errors.Is(err, combat.ErrNotFound) {
		return nil, err
	}
	active, err := characterInBattle(ctx, tx, in.CharacterID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, combat.ErrInBattle
	}
	if locationID != in.LocationID {
		return nil, world.ErrPositionChanged
	}
	location, err := scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1 FOR SHARE`, locationID))
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `SELECT id FROM planets WHERE id=$1 FOR SHARE`, location.PlanetID); err != nil {
		return nil, err
	}
	planet, err := worldPlanet(ctx, tx, location.PlanetID)
	if err != nil {
		return nil, err
	}
	if !world.Access(faction, *planet, true).CanTravel {
		return nil, world.ErrAccessDenied
	}
	var npc combat.Opponent
	var placement int
	npc.SpawnID, npc.HP, npc.HPCurrent = in.NPCSpawnID, stats.HP, stats.HP
	err = tx.QueryRow(ctx, `SELECT n.id,t.id,t.name,n.level,n.rank FROM npc_spawns s
 JOIN location_npcs n ON n.id=s.location_npc_id JOIN npc_templates t ON t.id=n.npc_template_id
 WHERE s.id=$1 AND n.location_id=$2 FOR SHARE OF n,t`, in.NPCSpawnID, locationID).Scan(&placement, &npc.TemplateID, &npc.Name, &npc.Level, &npc.Rank)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNPCUnavailable
	}
	if err != nil {
		return nil, err
	}
	if deriveRules {
		rules, rulesErr := combat.RulesFor(npc.Level, npc.Rank)
		if rulesErr != nil {
			return nil, rulesErr
		}
		stats.HP = rules.HP
	}
	var lockedPlacement int
	err = tx.QueryRow(ctx, `SELECT location_npc_id FROM npc_spawns WHERE id=$1 FOR UPDATE`, in.NPCSpawnID).Scan(&lockedPlacement)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNPCUnavailable
	}
	if err != nil {
		return nil, err
	}
	if lockedPlacement != placement {
		return nil, combat.ErrNPCUnavailable
	}
	// The statement after the lock sees the latest occupation and server time.
	var at time.Time
	var respawn *time.Time
	err = tx.QueryRow(ctx, `SELECT clock_timestamp(),respawn_at,
 EXISTS(SELECT 1 FROM battles WHERE npc_spawn_id=$1 AND status='active') FROM npc_spawns WHERE id=$1`, in.NPCSpawnID).Scan(&at, &respawn, &active)
	if err != nil {
		return nil, err
	}
	if active || respawn != nil && at.Before(*respawn) {
		return nil, combat.ErrNPCUnavailable
	}
	resources, err := lockCharacterResources(ctx, tx, in.CharacterID)
	if err != nil {
		return nil, err
	}
	hp, _ := character.RecoverResourceWithRemainder(resources.Health.Current, resources.Health.Maximum, resources.Health.Remainder, resources.Health.UpdatedAt, at, character.HealthRecoveryPeriod)
	if hp <= 0 {
		return nil, combat.ErrDefeated
	}
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	b, err := scanBattle(tx.QueryRow(ctx, `INSERT INTO battles(character_id,npc_spawn_id,location_id,start_key,
 npc_template_id,npc_name,npc_level,npc_rank,npc_hp,npc_hp_current,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$9,$10) RETURNING `+battleColumns,
		in.CharacterID, in.NPCSpawnID, locationID, in.Key, npc.TemplateID, npc.Name, npc.Level, npc.Rank, stats.HP, at))
	if err != nil {
		return nil, err
	}
	return &combat.Change{Battle: b}, nil
}
