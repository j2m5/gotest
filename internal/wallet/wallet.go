// Package wallet owns integer balances and server-authored currency operations.
package wallet

import (
	"context"
	"errors"
	"time"

	"gotest/internal/character"
)

const (
	Credits     = "credits"
	CartelCoins = "cartel_coins"
)

var (
	ErrInvalidInput      = errors.New("wallet: invalid input")
	ErrNotFound          = errors.New("wallet: owner or wallet not found")
	ErrInsufficientFunds = errors.New("wallet: insufficient funds")
	ErrOverflow          = errors.New("wallet: balance overflow")
	ErrOperationConflict = errors.New("wallet: operation key already has a different payload")
)

// Operation comes only from trusted server code after checking the source event.
// Amount is positive for grants, negative for spending. Source+Key identifies one
// event within the recipient wallet; cartel coins use the character's account.
type Operation struct {
	CharacterID              int
	Currency                 string
	Amount                   int64
	Source, Key, Description string
}

// Result is the original outcome even if the current balance has since changed.
type Result struct {
	ID                    int64
	Currency              string
	Amount, Before, After int64
	Description           string
	CreatedAt             time.Time
	Replayed              bool
}

type Repository interface {
	Apply(context.Context, Operation) (*Result, error)
	Balances(context.Context, character.Actor, int) (*Balances, error)
	History(context.Context, character.Actor, int, int64) (*History, error)
}

// JSON uses decimal strings to preserve all int64 values in browser clients.
type Balances struct {
	CharacterID int   `json:"character_id"`
	Credits     int64 `json:"credits,string"`
	CartelCoins int64 `json:"cartel_coins,string"`
}
type Entry struct {
	ID           int64     `json:"id,string"`
	Currency     string    `json:"currency"`
	Amount       int64     `json:"amount,string"`
	BalanceAfter int64     `json:"balance_after,string"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}
type History struct {
	Entries    []Entry `json:"entries"`
	NextCursor *string `json:"next_cursor"`
}
