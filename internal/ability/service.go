package ability

import (
	"context"

	"gotest/internal/character"
)

type Repository interface {
	List(context.Context, character.Actor, int) (*Loadout, error)
}
type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, actor character.Actor, characterID int) (*Loadout, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if actor.UserID <= 0 || int64(actor.UserID) > 1<<31-1 || actor.SessionHash == "" {
		return nil, character.ErrUnauthenticated
	}
	if characterID <= 0 || int64(characterID) > 1<<31-1 {
		return nil, ErrInvalidInput
	}
	return s.repository.List(ctx, actor, characterID)
}
