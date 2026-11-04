package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"gotest/internal/wallet"
)

func (r *WalletRepository) Apply(ctx context.Context, operation wallet.Operation) (*wallet.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := wallet.ValidateOperation(operation); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	result, err := r.ApplyInTx(ctx, tx, operation)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyInTx joins a trusted caller's transaction (for example XP plus money).
// Lock order: character, then wallet. It never commits; on ANY error the caller
// must roll back the entire transaction. No session is required for server rewards.
func (r *WalletRepository) ApplyInTx(ctx context.Context, tx pgx.Tx, operation wallet.Operation) (*wallet.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := wallet.ValidateOperation(operation); err != nil {
		return nil, err
	}
	var owner int
	err := tx.QueryRow(ctx, `SELECT user_id FROM characters WHERE id=$1 FOR UPDATE`, operation.CharacterID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, wallet.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var walletID, balance int64
	err = tx.QueryRow(ctx, `SELECT id,balance FROM wallets WHERE currency=$1 AND (character_id=$2 OR user_id=$3) FOR UPDATE`, operation.Currency, operation.CharacterID, owner).Scan(&walletID, &balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, wallet.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	result := &wallet.Result{Currency: operation.Currency}
	err = tx.QueryRow(ctx, `SELECT id,amount,balance_before,balance_after,description,created_at FROM wallet_entries WHERE wallet_id=$1 AND source=$2 AND operation_key=$3`, walletID, operation.Source, operation.Key).
		Scan(&result.ID, &result.Amount, &result.Before, &result.After, &result.Description, &result.CreatedAt)
	if err == nil {
		if result.Amount != operation.Amount || result.Description != operation.Description {
			return nil, wallet.ErrOperationConflict
		}
		result.Replayed = true
		return result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	after, err := wallet.ChangedBalance(balance, operation.Amount)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance=$2 WHERE id=$1`, walletID, after); err != nil {
		return nil, err
	}
	result.Amount, result.Before, result.After, result.Description = operation.Amount, balance, after, operation.Description
	err = tx.QueryRow(ctx, `INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description)
  VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at`, walletID, operation.Source, operation.Key, operation.Amount, balance, after, operation.Description).Scan(&result.ID, &result.CreatedAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}
