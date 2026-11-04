package wallet_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"gotest/internal/wallet"
)

func TestWalletAmountArithmetic(t *testing.T) {
	for _, tc := range []struct {
		balance, delta, want int64
		err                  error
	}{
		{0, 100, 100, nil}, {100, -40, 60, nil}, {100, -100, 0, nil},
		{math.MaxInt64 - 1, 1, math.MaxInt64, nil}, {math.MaxInt64, -math.MaxInt64, 0, nil},
		{0, -1, 0, wallet.ErrInsufficientFunds}, {math.MaxInt64, 1, 0, wallet.ErrOverflow},
		{-1, 1, 0, wallet.ErrInvalidInput}, {0, 0, 0, wallet.ErrInvalidInput}, {0, math.MinInt64, 0, wallet.ErrInvalidInput},
	} {
		got, err := wallet.ChangedBalance(tc.balance, tc.delta)
		if got != tc.want || !errors.Is(err, tc.err) {
			t.Fatalf("%d %+d = %d %v; want %d %v", tc.balance, tc.delta, got, err, tc.want, tc.err)
		}
	}
}

func TestWalletOperationValidation(t *testing.T) {
	valid := wallet.Operation{CharacterID: 1, Currency: wallet.Credits, Amount: 1, Source: "mission", Key: "mission:42/reward", Description: "Награда за задание"}
	if err := wallet.ValidateOperation(valid); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*wallet.Operation){
		func(o *wallet.Operation) { o.CharacterID = 0 }, func(o *wallet.Operation) { o.CharacterID = 1 << 31 },
		func(o *wallet.Operation) { o.Currency = "dollars" }, func(o *wallet.Operation) { o.Amount = 0 }, func(o *wallet.Operation) { o.Amount = math.MinInt64 },
		func(o *wallet.Operation) { o.Source = "" }, func(o *wallet.Operation) { o.Source = "Battle" }, func(o *wallet.Operation) { o.Source = strings.Repeat("a", 33) },
		func(o *wallet.Operation) { o.Key = "/bad" }, func(o *wallet.Operation) { o.Key = strings.Repeat("a", 129) }, func(o *wallet.Operation) { o.Key = "Ключ" },
		func(o *wallet.Operation) { o.Description = " " }, func(o *wallet.Operation) { o.Description = strings.Repeat("я", 201) }, func(o *wallet.Operation) { o.Description = "line\nbreak" },
		func(o *wallet.Operation) { o.Description = " trailing " }, func(o *wallet.Operation) { o.Description = string([]byte{255}) },
	} {
		bad := valid
		change(&bad)
		if !errors.Is(wallet.ValidateOperation(bad), wallet.ErrInvalidInput) {
			t.Fatalf("accepted invalid operation: %+v", bad)
		}
	}
	valid.Currency = wallet.CartelCoins
	valid.Amount = -1
	valid.Description = strings.Repeat("я", 200)
	if err := wallet.ValidateOperation(valid); err != nil {
		t.Fatal(err)
	}
}
