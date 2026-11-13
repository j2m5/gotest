package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"gotest/internal/combat"
)

// ProcessDue claims the character first, preserving the command lock order.
// The transaction is the lease: crashes roll back, and another process can retry
// immediately. A failed encounter gets a durable delay without moving its combat
// clock. True means effects OR a retry were committed; a deferred error carries
// diagnostics, but must not pause the worker's processing of other encounters.
func (r *CombatRepository) ProcessDue(ctx context.Context) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer rollbackCharacterTx(tx)
	var battleID int64
	var characterID int
	err = tx.QueryRow(ctx, `SELECT b.id,b.character_id FROM battles b
 JOIN battle_executions e ON e.battle_id=b.id JOIN characters c ON c.id=b.character_id
 WHERE b.status='active' AND e.next_action_at IS NOT NULL
 AND GREATEST(LEAST(e.next_action_at,e.expires_at),e.retry_at)<=statement_timestamp()
 ORDER BY GREATEST(LEAST(e.next_action_at,e.expires_at),e.retry_at),b.id
 LIMIT 1 FOR UPDATE OF c SKIP LOCKED`).Scan(&battleID, &characterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// A join may have seen an older execution snapshot before the character lock
	// became available. Recheck under that lock before acting or scheduling a retry.
	var due bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM battles b
 JOIN battle_executions e ON e.battle_id=b.id WHERE b.id=$1 AND b.status='active'
 AND e.next_action_at IS NOT NULL
 AND GREATEST(LEAST(e.next_action_at,e.expires_at),e.retry_at)<=statement_timestamp())`, battleID).Scan(&due)
	if err != nil || !due {
		return false, err
	}
	// PostgreSQL statement errors abort a transaction. A savepoint lets us undo
	// every effect while retaining the character claim for an atomic deferral.
	attempt, err := tx.Begin(ctx)
	if err != nil {
		return false, err
	}
	changed, cause := r.processDueBattle(ctx, attempt, battleID, characterID)
	if cause != nil {
		if err = attempt.Rollback(ctx); err != nil {
			return false, errors.Join(cause, err)
		}
		deferred := &combat.ProcessingError{BattleID: battleID, CharacterID: characterID, Err: cause}
		err = tx.QueryRow(ctx, `UPDATE battle_executions
 SET retry_count=LEAST(retry_count+1,6),
 retry_at=clock_timestamp()+LEAST(30,1<<retry_count)*interval '1 second'
 WHERE battle_id=$1 RETURNING retry_at`, battleID).Scan(&deferred.RetryAt)
		if err != nil {
			return false, errors.Join(cause, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return false, errors.Join(cause, err)
		}
		return true, deferred
	}
	if err = attempt.Commit(ctx); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	if changed {
		r.notifyCharacter(characterID)
	}
	return changed, nil
}

func (r *CombatRepository) processDueBattle(ctx context.Context, tx pgx.Tx, battleID int64, characterID int) (bool, error) {
	b, err := lockCombatBattle(ctx, tx, characterID, battleID)
	if err != nil {
		return false, err
	}
	before, err := loadExecution(ctx, tx, b)
	if err != nil {
		return false, err
	}
	at, err := resourceClock(ctx, tx)
	if err != nil {
		return false, err
	}
	f, events, err := combat.Advance(before, at)
	if err != nil {
		return false, err
	}
	if err = r.saveExecution(ctx, tx, before, &f, events); err != nil {
		return false, err
	}
	if len(events) == 0 {
		// Also recover metadata if the wall clock moved back after claiming. Do
		// not touch finished rows: their complete execution snapshot is immutable.
		_, err = tx.Exec(ctx, `UPDATE battle_executions SET retry_count=0,retry_at=NULL
 WHERE battle_id=$1 AND retry_count<>0 AND next_action_at IS NOT NULL`, battleID)
		if err != nil {
			return false, err
		}
	}
	return len(events) > 0, nil
}
