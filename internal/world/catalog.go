// Package world owns the galaxy catalog, locations and movement rules.
package world

import (
	"context"
	"gotest/internal/character"
)

type Faction struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Planet struct {
	ID                 int      `json:"id"`
	Code               string   `json:"code"`
	Name               string   `json:"name"`
	NameEN             string   `json:"name_en"`
	Description        *string  `json:"description"`
	CoreFaction        *Faction `json:"core_faction"`
	ControllingFaction *Faction `json:"controlling_faction"`
}

type Catalog struct {
	Planets []Planet `json:"planets"`
}

type Repository interface {
	Catalog(context.Context) (*Catalog, error)
	Navigation(context.Context, character.Actor, int) (*Navigation, error)
	Travel(context.Context, character.Actor, TravelInput) (*TravelResult, error)
	Locations(context.Context, character.Actor, int, int) (*LocationCatalog, error)
	Move(context.Context, character.Actor, MoveInput) (*TravelResult, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Catalog(ctx context.Context) (*Catalog, error) {
	return s.repository.Catalog(ctx)
}
