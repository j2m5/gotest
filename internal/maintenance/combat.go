package maintenance

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const oldBattleLogs = `b.status<>'active' AND b.ended_at<=$1
 AND NOT EXISTS(SELECT 1 FROM battle_log_archives a WHERE a.battle_id=b.id)`

func previewBattleLogs(ctx context.Context, pool *pgxpool.Pool, cutoff time.Time) (int64, int64, error) {
	var logs, events int64
	err := pool.QueryRow(ctx, `WITH eligible AS (
 SELECT b.id,e.event_sequence FROM battles b LEFT JOIN battle_executions e ON e.battle_id=b.id
 WHERE `+oldBattleLogs+`)
 SELECT (SELECT count(*) FROM eligible),
 (SELECT count(*) FROM battle_events v JOIN eligible e ON e.id=v.battle_id WHERE v.sequence<>e.event_sequence)`, cutoff).Scan(&logs, &events)
	return logs, events, err
}

// One encounter per transaction bounds work independently of the credential
// batch size. The final-event FK and archive trigger validate the retained result.
func clearBattleLog(ctx context.Context, pool *pgxpool.Pool, cutoff, clearedAt time.Time) (int64, int64, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	var id int64
	var finalSequence int
	err = tx.QueryRow(ctx, `SELECT b.id,e.event_sequence FROM battles b
 LEFT JOIN battle_executions e ON e.battle_id=b.id WHERE `+oldBattleLogs+`
 ORDER BY b.ended_at,b.id LIMIT 1 FOR UPDATE OF b SKIP LOCKED`, cutoff).Scan(&id, &finalSequence)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	// Recheck after the row lock: another cleanup may have committed while this
	// statement was taking its snapshot. Finished battle data itself is immutable.
	var archived bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM battle_log_archives WHERE battle_id=$1)`, id).Scan(&archived); err != nil {
		return 0, 0, err
	}
	if archived {
		// A concurrent pass already handled this candidate. As with skipped
		// locks, any remaining work can be picked up by a later run.
		return 0, 0, nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO battle_log_archives(battle_id,final_sequence,cleared_at) VALUES($1,$2,$3)`, id, finalSequence, clearedAt)
	if err != nil {
		return 0, 0, err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM battle_events WHERE battle_id=$1 AND sequence<>$2`, id, finalSequence)
	if err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return 1, tag.RowsAffected(), nil
}
