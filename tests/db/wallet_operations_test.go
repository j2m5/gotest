package db_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"gotest/internal/db"
	"gotest/internal/wallet"
)

func walletOperation(characterID int, currency string, amount int64, key string) wallet.Operation {
	return wallet.Operation{CharacterID: characterID, Currency: currency, Amount: amount, Source: "test", Key: key, Description: "Тестовая операция"}
}

func walletState(t *testing.T, ctx context.Context, s *testStorage, characterID int, currency string) (int64, int) {
	t.Helper()
	var balance int64
	var count int
	err := s.pool.QueryRow(ctx, `SELECT w.balance,(SELECT count(*) FROM wallet_entries WHERE wallet_id=w.id) FROM wallets w WHERE currency=$2 AND (character_id=$1 OR user_id=(SELECT user_id FROM characters WHERE id=$1))`, characterID, currency).Scan(&balance, &count)
	if err != nil {
		t.Fatal(err)
	}
	return balance, count
}

func TestWalletOperationsReplayOriginalResult(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewWalletRepository(s.pool)
	grant := walletOperation(1, wallet.Credits, 100, "one")
	first, err := r.Apply(ctx, grant)
	if err != nil || first.Before != 0 || first.After != 100 || first.Replayed {
		t.Fatalf("grant: %+v %v", first, err)
	}
	debit := walletOperation(1, wallet.Credits, -40, "two")
	paid, err := r.Apply(ctx, debit)
	if err != nil || paid.Before != 100 || paid.After != 60 {
		t.Fatalf("debit: %+v %v", paid, err)
	}
	replay, err := db.NewWalletRepository(s.pool).Apply(ctx, grant)
	if err != nil || !replay.Replayed {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	replay.Replayed = false
	if !reflect.DeepEqual(first, replay) {
		t.Fatalf("replay changed original result: %+v %+v", first, replay)
	}
	for _, change := range []func(*wallet.Operation){func(o *wallet.Operation) { o.Amount++ }, func(o *wallet.Operation) { o.Description = "Другой результат" }} {
		conflict := grant
		change(&conflict)
		if _, err = r.Apply(ctx, conflict); !errors.Is(err, wallet.ErrOperationConflict) {
			t.Fatalf("payload conflict: %v", err)
		}
	}
	if balance, count := walletState(t, ctx, s, 1, wallet.Credits); balance != 60 || count != 2 {
		t.Fatalf("replay mutated wallet: %d/%d", balance, count)
	}
}

func TestWalletOperationsRejectInvalidInsufficientOverflowAndRollback(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewWalletRepository(s.pool)
	if _, err := r.Apply(ctx, walletOperation(1, wallet.Credits, -1, "poor")); !errors.Is(err, wallet.ErrInsufficientFunds) {
		t.Fatalf("debit zero: %v", err)
	}
	if _, err := r.Apply(ctx, walletOperation(1, wallet.Credits, math.MaxInt64, "max")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Apply(ctx, walletOperation(1, wallet.Credits, 1, "overflow")); !errors.Is(err, wallet.ErrOverflow) {
		t.Fatalf("overflow: %v", err)
	}
	if balance, count := walletState(t, ctx, s, 1, wallet.Credits); balance != math.MaxInt64 || count != 1 {
		t.Fatalf("overflow changed state %d/%d", balance, count)
	}
	for _, op := range []wallet.Operation{walletOperation(1, wallet.Credits, 0, "bad"), walletOperation(0, wallet.Credits, 1, "bad"), walletOperation(1, "other", 1, "bad")} {
		if _, err := r.Apply(ctx, op); !errors.Is(err, wallet.ErrInvalidInput) {
			t.Fatalf("invalid operation: %v", err)
		}
	}
	if _, err := r.Apply(ctx, walletOperation(999999, wallet.Credits, 1, "missing")); !errors.Is(err, wallet.ErrNotFound) {
		t.Fatalf("missing character: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := r.Apply(cancelled, walletOperation(2, wallet.Credits, 1, "cancelled")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	mustExec(t, ctx, s, `CREATE FUNCTION fail_wallet_entry() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'test ledger failure'; END $$;
 CREATE TRIGGER fail_wallet_entry BEFORE INSERT ON wallet_entries FOR EACH ROW EXECUTE FUNCTION fail_wallet_entry()`)
	if _, err := r.Apply(ctx, walletOperation(2, wallet.Credits, 10, "rollback")); err == nil {
		t.Fatal("ledger failure not reported")
	}
	if balance, count := walletState(t, ctx, s, 2, wallet.Credits); balance != 0 || count != 0 {
		t.Fatalf("partial commit: %d/%d", balance, count)
	}
}

func TestWalletConcurrentSharedAccountOperations(t *testing.T) {
	for _, mode := range []string{"same_key", "different_keys", "overspend"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			r := db.NewWalletRepository(s.pool)
			if mode == "overspend" {
				if _, err := r.Apply(ctx, walletOperation(1, wallet.CartelCoins, 150, "initial")); err != nil {
					t.Fatal(err)
				}
			}
			type outcome struct {
				result *wallet.Result
				err    error
			}
			const workers = 10
			start, done := make(chan struct{}), make(chan outcome, workers)
			for i := range workers {
				go func() {
					<-start
					key := fmt.Sprintf("operation:%d", i)
					amount := int64(10)
					if mode == "same_key" {
						key = "shared"
					}
					if mode == "overspend" {
						amount = -100
					}
					result, err := r.Apply(ctx, walletOperation(1+i%2, wallet.CartelCoins, amount, key))
					done <- outcome{result, err}
				}()
			}
			close(start)
			fresh, insufficient := 0, 0
			for range workers {
				got := <-done
				if errors.Is(got.err, wallet.ErrInsufficientFunds) {
					insufficient++
					continue
				}
				if got.err != nil {
					t.Fatal(got.err)
				}
				if !got.result.Replayed {
					fresh++
				}
			}
			wantBalance, wantCount, wantFresh, wantInsufficient := int64(100), 10, 10, 0
			if mode == "same_key" {
				wantBalance, wantCount, wantFresh = 10, 1, 1
			}
			if mode == "overspend" {
				wantBalance, wantCount, wantFresh, wantInsufficient = 50, 2, 1, 9
			}
			if balance, count := walletState(t, ctx, s, 2, wallet.CartelCoins); balance != wantBalance || count != wantCount || fresh != wantFresh || insufficient != wantInsufficient {
				t.Fatalf("concurrent result %d/%d fresh=%d insufficient=%d", balance, count, fresh, insufficient)
			}
			if balance, count := walletState(t, ctx, s, 3, wallet.CartelCoins); balance != 0 || count != 0 {
				t.Fatalf("another account changed %d/%d", balance, count)
			}
		})
	}
}
