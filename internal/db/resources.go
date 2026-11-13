package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
)

func lockCharacterResources(ctx context.Context, tx pgx.Tx, characterID int) (character.Resources, error) {
	var resources character.Resources
	err := tx.QueryRow(ctx, `SELECT hp_current,hp,hp_regen_remainder,hp_updated_at,
	 mp_current,mp,mp_regen_remainder,mp_updated_at FROM characters WHERE id=$1 FOR UPDATE`, characterID).Scan(
		&resources.Health.Current, &resources.Health.Maximum, &resources.Health.Remainder, &resources.Health.UpdatedAt,
		&resources.Energy.Current, &resources.Energy.Maximum, &resources.Energy.Remainder, &resources.Energy.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = character.ErrNotFound
	}
	return resources, err
}

func resourceClock(ctx context.Context, tx pgx.Tx) (time.Time, error) {
	var at time.Time
	err := tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&at)
	return at, err
}

func saveCharacterResources(ctx context.Context, tx pgx.Tx, characterID int, resources character.Resources) error {
	_, err := tx.Exec(ctx, `UPDATE characters SET hp_current=$2,hp_regen_remainder=$3,hp_updated_at=$4,
	 mp_current=$5,mp_regen_remainder=$6,mp_updated_at=$7,resources_revision=resources_revision+1 WHERE id=$1`,
		characterID, resources.Health.Current, resources.Health.Remainder, resources.Health.UpdatedAt,
		resources.Energy.Current, resources.Energy.Remainder, resources.Energy.UpdatedAt)
	return err
}

// ChangeCharacterResourcesInTx is for trusted server operations. The caller
// authorizes the action and owns the transaction; roll back on ANY error. Do not
// expose arbitrary deltas through HTTP. Lock order starts with the character.
func ChangeCharacterResourcesInTx(ctx context.Context, tx pgx.Tx, characterID int, change character.ResourceChange) (*character.Character, error) {
	if characterID <= 0 || int64(characterID) > 1<<31-1 {
		return nil, character.ErrNotFound
	}
	resources, err := lockCharacterResources(ctx, tx, characterID)
	if err != nil {
		return nil, err
	}
	at, err := resourceClock(ctx, tx)
	if err != nil {
		return nil, err
	}
	next, err := character.ChangeResources(resources, change, at)
	if err != nil {
		return nil, err
	}
	if next != resources {
		if err = saveCharacterResources(ctx, tx, characterID, next); err != nil {
			return nil, err
		}
	}
	return scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1`, characterID))
}
