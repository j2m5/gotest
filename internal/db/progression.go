package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/progression"
)

type ProgressionRepository struct{ pool *pgxpool.Pool }

func NewProgressionRepository(pool *pgxpool.Pool) *ProgressionRepository {
	return &ProgressionRepository{pool: pool}
}

var _ progression.Repository = (*ProgressionRepository)(nil)

func (r *ProgressionRepository) Grant(ctx context.Context, reward progression.Reward) (*progression.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := progression.ValidateReward(reward); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()

	// Serialize rewards for this character, including competing first attempts
	// with the same event key. No account/session lock is needed for server rewards.
	var current progression.State
	err = tx.QueryRow(ctx, `SELECT level,experience FROM characters WHERE id=$1 FOR UPDATE`, reward.CharacterID).
		Scan(&current.Level, &current.Experience)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, progression.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	result := &progression.Result{CharacterID: reward.CharacterID, Key: reward.Key, Requested: reward.Experience}
	err = tx.QueryRow(ctx, `SELECT requested_amount,applied_amount,level_before,experience_before,level_after,experience_after
		FROM character_experience_rewards WHERE character_id=$1 AND reward_key=$2`, reward.CharacterID, reward.Key).
		Scan(&result.Requested, &result.Change.Granted, &result.Change.Before.Level, &result.Change.Before.Experience,
			&result.Change.After.Level, &result.Change.After.Experience)
	if err == nil {
		if result.Requested != reward.Experience {
			return nil, progression.ErrRewardConflict
		}
		result.Replayed = true
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// One query sees one coherent catalog snapshot; future grants read new balance
	// values. Replays above do not depend on today's curve being valid or unchanged.
	rows, err := tx.Query(ctx, `SELECT level,experience_required FROM levels ORDER BY level`)
	if err != nil {
		return nil, err
	}
	levels, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (progression.Level, error) {
		var level progression.Level
		err := row.Scan(&level.Number, &level.ExperienceRequired)
		return level, err
	})
	if err != nil {
		return nil, err
	}
	curve, err := progression.NewCurve(levels)
	if err != nil {
		return nil, err
	}
	result.Change, err = curve.Advance(current, reward.Experience)
	if err != nil {
		return nil, err
	}
	if result.Change.After != current {
		_, err = tx.Exec(ctx, `UPDATE characters SET level=$2,experience=$3 WHERE id=$1`,
			reward.CharacterID, result.Change.After.Level, result.Change.After.Experience)
		if err != nil {
			return nil, err
		}
	}
	// Record even a zero award at the cap: adding levels later must not revive it.
	_, err = tx.Exec(ctx, `INSERT INTO character_experience_rewards
		(character_id,reward_key,requested_amount,applied_amount,level_before,experience_before,level_after,experience_after)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, reward.CharacterID, reward.Key, result.Requested, result.Change.Granted,
		result.Change.Before.Level, result.Change.Before.Experience, result.Change.After.Level, result.Change.After.Experience)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
