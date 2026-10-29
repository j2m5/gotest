package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
)

func (r *CharacterRepository) Profile(ctx context.Context, actor character.Actor, id int) (*character.PublicProfile, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err := authenticateCharacterActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	profile := &character.PublicProfile{}
	err = tx.QueryRow(ctx, `SELECT c.id,c.name,c.faction_id,c.class_id,c.level,c.avatar,role.name
		FROM characters c JOIN users u ON u.id=c.user_id
		JOIN account_roles role ON role.id=u.role_id WHERE c.id=$1`, id).Scan(
		&profile.ID, &profile.Name, &profile.FactionID, &profile.ClassID, &profile.Level, &profile.Avatar, &profile.AccountRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, character.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return profile, nil
}
