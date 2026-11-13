package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
)

func appendCombatEvents(ctx context.Context, tx pgx.Tx, battleID int64, events []combat.Event) error {
	if len(events) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}
		batch.Queue(`INSERT INTO battle_events(battle_id,sequence,occurred_at,payload) VALUES($1,$2,$3,$4)`, battleID, event.Sequence, event.OccurredAt, payload)
	}
	return tx.SendBatch(ctx, batch).Close()
}

func (r *CombatRepository) View(ctx context.Context, actor character.Actor, battleID int64, characterID, after int) (*combat.View, error) {
	if !combat.ValidID(characterID) || battleID <= 0 || after < 0 || int64(after) > 1<<31-1 {
		return nil, combat.ErrInvalidInput
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	if err = requireSelectedCharacter(ctx, tx, actor, characterID); err != nil {
		return nil, err
	}
	b, err := scanBattle(tx.QueryRow(ctx, `SELECT `+battleColumns+` FROM battles WHERE id=$1 AND character_id=$2`, battleID, characterID))
	if err != nil {
		return nil, err
	}
	c, err := scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1 AND user_id=$2`, characterID, actor.UserID))
	if err != nil {
		return nil, err
	}
	v := &combat.View{State: combat.State{Character: c, Battle: b}, Events: []combat.Event{}}
	var finalSequence *int
	err = tx.QueryRow(ctx, `SELECT e.next_action_at,e.expires_at,a.cleared_at,a.final_sequence
 FROM battle_executions e LEFT JOIN battle_log_archives a USING(battle_id) WHERE e.battle_id=$1`, battleID).
		Scan(&v.NextActionAt, &v.ExpiresAt, &v.LogClearedAt, &finalSequence)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	// An archived response replaces the client's journal, including a client that
	// already read past the final event or saw a previous page before cleanup.
	rows, err := tx.Query(ctx, `SELECT payload FROM battle_events WHERE battle_id=$1
 AND (($3::integer IS NULL AND sequence>$2) OR sequence=$3) ORDER BY sequence LIMIT 51`, battleID, after, finalSequence)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var raw []byte
		var event combat.Event
		if err = rows.Scan(&raw); err == nil {
			err = json.Unmarshal(raw, &event)
		}
		if err != nil {
			rows.Close()
			return nil, err
		}
		v.Events = append(v.Events, event)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(v.Events) > 50 {
		cursor := v.Events[49].Sequence
		v.NextAfter = &cursor
		v.Events = v.Events[:50]
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return v, nil
}
