package world

import (
	"context"
	"errors"

	"gotest/internal/character"
)

var (
	ErrWrongPlanet      = errors.New("world: location is outside the current planet")
	ErrLocationNotFound = errors.New("world: location not found")
)

type LocationCatalog struct {
	Position       Position   `json:"position"`
	Locations      []Location `json:"locations"`
	ReturnRequired bool       `json:"return_required"`
}

type MoveInput struct {
	CharacterID    int `json:"character_id"`
	LocationID     int `json:"location_id"`
	FromLocationID int `json:"from_location_id"`
}

func (s *Service) Locations(ctx context.Context, actor character.Actor, characterID, planetID int) (*LocationCatalog, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	if !validID(characterID) || !validID(planetID) {
		return nil, ErrInvalidInput
	}
	return s.repository.Locations(ctx, actor, characterID, planetID)
}

func (s *Service) Move(ctx context.Context, actor character.Actor, in MoveInput) (*TravelResult, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	if !validID(in.CharacterID) || !validID(in.LocationID) || !validID(in.FromLocationID) {
		return nil, ErrInvalidInput
	}
	return s.repository.Move(ctx, actor, in)
}
