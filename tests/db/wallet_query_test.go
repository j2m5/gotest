package db_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/wallet"
)

func TestWalletReadsSeparateCreditsAndShareCartelCoins(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewWalletRepository(s.pool)
	service := wallet.NewService(r)
	a, b := repositoryActor(t, ctx, s, 1, "wallet-first"), repositoryActor(t, ctx, s, 2, "wallet-other")
	mustExec(t, ctx, s, `UPDATE users SET character_id=CASE id WHEN 1 THEN 1 ELSE 3 END`)
	for _, op := range []wallet.Operation{walletOperation(1, wallet.Credits, 100, "one"), walletOperation(2, wallet.Credits, 20, "two"), walletOperation(1, wallet.CartelCoins, 5, "three"), walletOperation(2, wallet.CartelCoins, 6, "four")} {
		if _, err := r.Apply(ctx, op); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM wallets t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM wallet_entries t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM characters t))::text`
	var before, after string
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	first, err := service.Balances(ctx, a, 1)
	if err != nil || first.Credits != 100 || first.CartelCoins != 11 {
		t.Fatalf("first wallet %+v %v", first, err)
	}
	history, err := service.History(ctx, a, 1, 0)
	if err != nil || len(history.Entries) != 3 || history.NextCursor != nil {
		t.Fatalf("first history %+v %v", history, err)
	}
	if history.Entries[0].Currency != wallet.CartelCoins || history.Entries[0].Amount != 6 {
		t.Fatal("shared account history missing")
	}
	if err = s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
		t.Fatal("wallet read changed data")
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1`)
	second, err := service.Balances(ctx, a, 2)
	if err != nil || second.Credits != 20 || second.CartelCoins != 11 {
		t.Fatalf("second wallet %+v %v", second, err)
	}
	other, err := service.Balances(ctx, b, 3)
	if err != nil || other.Credits != 0 || other.CartelCoins != 0 {
		t.Fatalf("other wallet %+v %v", other, err)
	}
	empty, err := service.History(ctx, b, 3, 0)
	if err != nil || empty.Entries == nil || len(empty.Entries) != 0 || empty.NextCursor != nil {
		t.Fatalf("other history %+v %v", empty, err)
	}
}

func TestWalletHistoryPaginatesWithoutLosingIntegerPrecision(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewWalletRepository(s.pool)
	service := wallet.NewService(r)
	actor := repositoryActor(t, ctx, s, 1, "wallet-paging")
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1; ALTER SEQUENCE wallet_entries_id_seq RESTART WITH 9007199254740993`)
	for i := range 23 {
		if _, err := r.Apply(ctx, walletOperation(1, wallet.Credits, 1, fmt.Sprintf("page:%d", i))); err != nil {
			t.Fatal(err)
		}
	}
	first, err := service.History(ctx, actor, 1, 0)
	if err != nil || len(first.Entries) != 20 || first.NextCursor == nil {
		t.Fatalf("first page %+v %v", first, err)
	}
	cursor, err := strconv.ParseInt(*first.NextCursor, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.History(ctx, actor, 1, cursor)
	if err != nil || len(second.Entries) != 3 || second.NextCursor != nil {
		t.Fatalf("second page %+v %v", second, err)
	}
	seen := map[int64]bool{}
	for _, page := range []*wallet.History{first, second} {
		for _, entry := range page.Entries {
			if seen[entry.ID] || entry.ID < 9007199254740993 {
				t.Fatalf("invalid/repeated ID %d", entry.ID)
			}
			seen[entry.ID] = true
		}
	}
	if len(seen) != 23 {
		t.Fatal("paging lost entries")
	}
}

func TestWalletReadsRequireSelectedAuthenticatedCharacter(t *testing.T) {
	for _, change := range []string{"other_character", "no_selection", "other_account", "expired", "revoked", "ban"} {
		t.Run(change, func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			service := wallet.NewService(db.NewWalletRepository(s.pool))
			actor := repositoryActor(t, ctx, s, 1, "wallet-guard")
			mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
			want := character.ErrUnauthenticated
			switch change {
			case "other_character":
				mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1`)
				want = character.ErrSelectionChanged
			case "no_selection":
				mustExec(t, ctx, s, `UPDATE users SET character_id=NULL WHERE id=1`)
				want = character.ErrSelectionRequired
			case "other_account":
				actor.UserID = 2
			case "expired":
				mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
			case "revoked":
				mustExec(t, ctx, s, `DELETE FROM sessions WHERE user_id=1`)
			case "ban":
				mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','test',2)`)
			}
			if _, err := service.Balances(ctx, actor, 1); !errors.Is(err, want) {
				t.Fatalf("balance guard %s: %v", change, err)
			}
			if _, err := service.History(ctx, actor, 1, 0); !errors.Is(err, want) {
				t.Fatalf("history guard %s: %v", change, err)
			}
		})
	}
}
