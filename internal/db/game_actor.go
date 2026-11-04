package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
)

// Presence and world actions share account-wide session and ban checks.
func authenticateGameActor(ctx context.Context, tx pgx.Tx, actor character.Actor, lock bool) error {
	if err := authenticateCharacterActor(ctx, tx, actor, lock); err != nil {
		return err
	}
	var banned bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id=$1 AND scope='account'
	 AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at>clock_timestamp()))`, actor.UserID).Scan(&banned)
	if err != nil {
		return err
	}
	if banned {
		return character.ErrUnauthenticated
	}
	return nil
}

// Called after authentication in the same snapshot/transaction. The database's
// composite FK guarantees that users.character_id belongs to that account.
func requireSelectedCharacter(ctx context.Context, tx pgx.Tx, actor character.Actor, expected int) error {
	var selected *int
	if err := tx.QueryRow(ctx, `SELECT character_id FROM users WHERE id=$1`, actor.UserID).Scan(&selected); err != nil {
		return err
	}
	if selected == nil {
		return character.ErrSelectionRequired
	}
	if *selected != expected {
		return character.ErrSelectionChanged
	}
	return nil
}
