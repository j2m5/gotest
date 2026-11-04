package routes_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/wallet"
)

type walletStub struct {
	empty bool
	err   error
}

func (s *walletStub) Apply(context.Context, wallet.Operation) (*wallet.Result, error) {
	return nil, errors.New("HTTP must not apply currency operations")
}
func (s *walletStub) Balances(_ context.Context, _ character.Actor, id int) (*wallet.Balances, error) {
	return &wallet.Balances{CharacterID: id, Credits: math.MaxInt64, CartelCoins: 0}, s.err
}
func (s *walletStub) History(context.Context, character.Actor, int, int64) (*wallet.History, error) {
	result := &wallet.History{Entries: []wallet.Entry{}}
	if !s.empty {
		cursor := "9007199254740993"
		result.NextCursor = &cursor
		result.Entries = append(result.Entries, wallet.Entry{ID: 9007199254740993, Currency: wallet.Credits, Amount: -123, BalanceAfter: math.MaxInt64 - 123, Description: "Покупка припасов", CreatedAt: time.Now().UTC()})
	}
	return result, s.err
}

func TestWalletRoutesAreReadOnlyAndReportUnavailability(t *testing.T) {
	h := router(t, &authStub{user: &auth.Identity{ID: 1}}, false, 30)
	for _, path := range []string{"/api/wallet?character_id=1", "/api/wallet/history?character_id=1"} {
		w := request(h, "POST", path, `{"amount":100}`, characterCookie())
		requireAPIError(t, w, 405, "method_not_allowed")
		if w.Header().Get("Allow") != "GET" {
			t.Fatalf("wallet methods: %s", w.Header().Get("Allow"))
		}
		requireAPIError(t, request(h, "GET", path, "", characterCookie()), 503, "not_ready")
	}
}
