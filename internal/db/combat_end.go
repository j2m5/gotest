package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
)

// lockCombatBattle is shared by player commands and the future NPC worker.
// A worker must never acquire an account lock after entering this sequence.
func lockCombatBattle(ctx context.Context, tx pgx.Tx, characterID int, battleID int64) (*combat.Battle, error) {
	var id int
	err := tx.QueryRow(ctx, `SELECT id FROM characters WHERE id=$1 FOR UPDATE`, characterID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var spawnID int
	err = tx.QueryRow(ctx, `SELECT npc_spawn_id FROM battles WHERE id=$1 AND character_id=$2`, battleID, characterID).Scan(&spawnID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err = tx.QueryRow(ctx, `SELECT id FROM npc_spawns WHERE id=$1 FOR UPDATE`, spawnID).Scan(&id); err != nil {
		return nil, err
	}
	return scanBattle(tx.QueryRow(ctx, `SELECT `+battleColumns+` FROM battles WHERE id=$1 AND character_id=$2 FOR UPDATE`, battleID, characterID))
}

// RetreatInTx does not check current planet access: losing a planet must never
// trap a character in battle. The caller commits the ending and its event together.
// Resolve already-due NPC actions first when the combat executor is connected.
func (r *CombatRepository) RetreatInTx(ctx context.Context, tx pgx.Tx, actor character.Actor, characterID int, battleID int64) (*combat.Change, error) {
	if !combat.ValidID(characterID) || battleID <= 0 {
		return nil, combat.ErrInvalidInput
	}
	if err := authenticateGameActor(ctx, tx, actor, true); err != nil {
		return nil, err
	}
	if _, _, err := worldCharacter(ctx, tx, actor, characterID, true); err != nil {
		return nil, err
	}
	b, err := lockCombatBattle(ctx, tx, characterID, battleID)
	if err != nil {
		return nil, err
	}
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	if b.Status != combat.Active {
		return &combat.Change{Battle: b, Replayed: true}, nil
	}
	return releaseCombatBattle(ctx, tx, b, combat.Retreated)
}

// DefeatInTx is a trusted executor operation, not an HTTP command. Its caller has
// resolved a lethal effect in the same transaction. It remains valid after logout
// or character switching. Roll back the entire action on any error.
func (r *CombatRepository) DefeatInTx(ctx context.Context, tx pgx.Tx, characterID int, battleID int64) (*combat.Change, error) {
	if !combat.ValidID(characterID) || battleID <= 0 {
		return nil, combat.ErrInvalidInput
	}
	b, err := lockCombatBattle(ctx, tx, characterID, battleID)
	if err != nil {
		return nil, err
	}
	if b.Status != combat.Active {
		return &combat.Change{Battle: b, Replayed: true}, nil
	}
	resources, err := lockCharacterResources(ctx, tx, characterID)
	if err != nil {
		return nil, err
	}
	if _, err = ChangeCharacterResourcesInTx(ctx, tx, characterID, character.ResourceChange{HP: -resources.Health.Maximum}); err != nil {
		return nil, err
	}
	return releaseCombatBattle(ctx, tx, b, combat.Defeated)
}

func releaseCombatBattle(ctx context.Context, tx pgx.Tx, b *combat.Battle, status combat.Status) (*combat.Change, error) {
	ended, err := scanBattle(tx.QueryRow(ctx, `UPDATE battles SET status=$2,npc_hp_current=npc_hp,
 ended_at=GREATEST(clock_timestamp(),created_at) WHERE id=$1 RETURNING `+battleColumns, b.ID, status))
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE npc_spawns SET respawn_at=NULL WHERE id=$1`, b.NPC.SpawnID); err != nil {
		return nil, err
	}
	return &combat.Change{Battle: ended}, nil
}
