package db_test

import (
	"errors"
	"testing"

	"gotest/internal/db"
	"gotest/internal/progression"
	"gotest/internal/wallet"
)

func TestWalletAndExperienceShareTransaction(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	money, xp := db.NewWalletRepository(s.pool), db.NewProgressionRepository(s.pool)
	reward := progression.Reward{CharacterID: 1, Key: "battle:atomic", Experience: 10}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = xp.GrantInTx(ctx, tx, reward); err != nil {
		t.Fatal(err)
	}
	if _, err = money.ApplyInTx(ctx, tx, walletOperation(1, wallet.Credits, -1, "atomic")); !errors.Is(err, wallet.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds: %v", err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 1, Experience: 0})
	assertRewardCount(t, ctx, s, 0)
	if balance, count := walletState(t, ctx, s, 1, wallet.Credits); balance != 0 || count != 0 {
		t.Fatal("currency partially committed")
	}
	for pass := range 2 {
		tx, err = s.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		experience, err := xp.GrantInTx(ctx, tx, reward)
		if err != nil {
			t.Fatal(err)
		}
		currency, err := money.ApplyInTx(ctx, tx, walletOperation(1, wallet.Credits, 50, "atomic"))
		if err != nil {
			t.Fatal(err)
		}
		if experience.Replayed != (pass == 1) || currency.Replayed != (pass == 1) {
			t.Fatal("combined retry did not replay both outcomes")
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 1, Experience: 10})
	assertRewardCount(t, ctx, s, 1)
	if balance, count := walletState(t, ctx, s, 1, wallet.Credits); balance != 50 || count != 1 {
		t.Fatalf("combined retry changed money: %d/%d", balance, count)
	}
}

func TestWalletLedgerFailureRollsBackExperience(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `CREATE FUNCTION fail_wallet_reward() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'test ledger failure'; END $$;
 CREATE TRIGGER fail_wallet_reward BEFORE INSERT ON wallet_entries FOR EACH ROW EXECUTE FUNCTION fail_wallet_reward()`)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = db.NewProgressionRepository(s.pool).GrantInTx(ctx, tx, progression.Reward{CharacterID: 1, Key: "battle:failed", Experience: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.NewWalletRepository(s.pool).ApplyInTx(ctx, tx, walletOperation(1, wallet.Credits, 50, "failed")); err == nil {
		t.Fatal("expected ledger failure")
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertProgressionState(t, ctx, s, 1, progression.State{Level: 1, Experience: 0})
	assertRewardCount(t, ctx, s, 0)
	if balance, count := walletState(t, ctx, s, 1, wallet.Credits); balance != 0 || count != 0 {
		t.Fatal("failed combined grant changed money")
	}
}
