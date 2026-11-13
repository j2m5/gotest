package combat

import (
	"context"
	"gotest/internal/character"
)

type Repository interface {
	Current(context.Context, character.Actor, int) (*State, error)
	Start(context.Context, character.Actor, StartInput) (*Result, error)
	Use(context.Context, character.Actor, int64, UseInput) (*Result, error)
	Retreat(context.Context, character.Actor, int64, int) (*Result, error)
	View(context.Context, character.Actor, int64, int, int) (*View, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Current(ctx context.Context, actor character.Actor, characterID int) (*State, error) {
	if err := validateActor(ctx, actor, characterID); err != nil {
		return nil, err
	}
	return s.repository.Current(ctx, actor, characterID)
}

func validateActor(ctx context.Context, actor character.Actor, characterID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !ValidID(actor.UserID) || actor.SessionHash == "" {
		return character.ErrUnauthenticated
	}
	if !ValidID(characterID) {
		return ErrInvalidInput
	}
	return nil
}

func (s *Service) Start(ctx context.Context, actor character.Actor, in StartInput) (*Result, error) {
	if err := validateActor(ctx, actor, in.CharacterID); err != nil {
		return nil, err
	}
	if err := ValidateStart(in, NPCStats{HP: 1}); err != nil {
		return nil, err
	}
	return s.repository.Start(ctx, actor, in)
}
func (s *Service) Use(ctx context.Context, actor character.Actor, id int64, in UseInput) (*Result, error) {
	if err := validateActor(ctx, actor, in.CharacterID); err != nil {
		return nil, err
	}
	if err := ValidateUse(id, in); err != nil {
		return nil, err
	}
	return s.repository.Use(ctx, actor, id, in)
}
func (s *Service) Retreat(ctx context.Context, actor character.Actor, id int64, characterID int) (*Result, error) {
	if err := validateActor(ctx, actor, characterID); err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repository.Retreat(ctx, actor, id, characterID)
}
func (s *Service) View(ctx context.Context, actor character.Actor, id int64, characterID, after int) (*View, error) {
	if err := validateActor(ctx, actor, characterID); err != nil {
		return nil, err
	}
	if id <= 0 || after < 0 || int64(after) > 1<<31-1 {
		return nil, ErrInvalidInput
	}
	return s.repository.View(ctx, actor, id, characterID, after)
}
