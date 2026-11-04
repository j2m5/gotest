package db

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
	"gotest/internal/wallet"
)

type WalletRepository struct{ pool *pgxpool.Pool }

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository { return &WalletRepository{pool: pool} }

var _ wallet.Repository = (*WalletRepository)(nil)

func (r *WalletRepository) readTransaction(ctx context.Context, actor character.Actor, expected int) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	if err = authenticateGameActor(ctx, tx, actor, false); err == nil {
		err = requireSelectedCharacter(ctx, tx, actor, expected)
	}
	if err != nil {
		rollbackCharacterTx(tx)
		return nil, err
	}
	return tx, nil
}

func (r *WalletRepository) Balances(ctx context.Context, actor character.Actor, expected int) (*wallet.Balances, error) {
	tx, err := r.readTransaction(ctx, actor, expected)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	result := &wallet.Balances{CharacterID: expected}
	err = tx.QueryRow(ctx, `SELECT c.balance,p.balance FROM wallets c JOIN wallets p ON p.user_id=$1 AND p.currency='cartel_coins'
 WHERE c.character_id=$2 AND c.currency='credits'`, actor.UserID, expected).Scan(&result.Credits, &result.CartelCoins)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *WalletRepository) History(ctx context.Context, actor character.Actor, expected int, before int64) (*wallet.History, error) {
	tx, err := r.readTransaction(ctx, actor, expected)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	result := &wallet.History{Entries: []wallet.Entry{}}
	rows, err := tx.Query(ctx, `SELECT e.id,w.currency,e.amount,e.balance_after,e.description,e.created_at
 FROM wallet_entries e JOIN wallets w ON w.id=e.wallet_id
 WHERE (w.character_id=$1 OR w.user_id=$2) AND ($3::BIGINT=0 OR e.id<$3)
 ORDER BY e.id DESC LIMIT 21`, expected, actor.UserID, before)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var entry wallet.Entry
		if err = rows.Scan(&entry.ID, &entry.Currency, &entry.Amount, &entry.BalanceAfter, &entry.Description, &entry.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		result.Entries = append(result.Entries, entry)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(result.Entries) > 20 {
		result.Entries = result.Entries[:20]
		cursor := strconv.FormatInt(result.Entries[19].ID, 10)
		result.NextCursor = &cursor
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
