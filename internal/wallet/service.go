package wallet

import (
	"context"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"gotest/internal/character"
)

var sourcePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9:._/-]{0,127}$`)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Apply(ctx context.Context, operation Operation) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ValidateOperation(operation); err != nil {
		return nil, err
	}
	return s.repository.Apply(ctx, operation)
}

func validID(id int) bool { return id > 0 && int64(id) <= math.MaxInt32 }

func validateRead(ctx context.Context, actor character.Actor, characterID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(actor.UserID) || actor.SessionHash == "" {
		return character.ErrUnauthenticated
	}
	if !validID(characterID) {
		return ErrInvalidInput
	}
	return nil
}
func (s *Service) Balances(ctx context.Context, actor character.Actor, characterID int) (*Balances, error) {
	if err := validateRead(ctx, actor, characterID); err != nil {
		return nil, err
	}
	return s.repository.Balances(ctx, actor, characterID)
}
func (s *Service) History(ctx context.Context, actor character.Actor, characterID int, before int64) (*History, error) {
	if err := validateRead(ctx, actor, characterID); err != nil {
		return nil, err
	}
	if before < 0 {
		return nil, ErrInvalidInput
	}
	return s.repository.History(ctx, actor, characterID, before)
}

func ValidateOperation(operation Operation) error {
	if !validID(operation.CharacterID) || (operation.Currency != Credits && operation.Currency != CartelCoins) ||
		operation.Amount == 0 || operation.Amount == math.MinInt64 || !sourcePattern.MatchString(operation.Source) || !keyPattern.MatchString(operation.Key) {
		return ErrInvalidInput
	}
	text := operation.Description
	if !utf8.ValidString(text) || text == "" || strings.TrimSpace(text) != text || utf8.RuneCountInString(text) > 200 {
		return ErrInvalidInput
	}
	for _, r := range text {
		if unicode.IsControl(r) {
			return ErrInvalidInput
		}
	}
	return nil
}

func ChangedBalance(balance, amount int64) (int64, error) {
	if balance < 0 || amount == 0 || amount == math.MinInt64 {
		return 0, ErrInvalidInput
	}
	if amount < 0 && balance < -amount {
		return 0, ErrInsufficientFunds
	}
	if amount > 0 && balance > math.MaxInt64-amount {
		return 0, ErrOverflow
	}
	return balance + amount, nil
}
