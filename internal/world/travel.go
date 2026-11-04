package world

import (
	"context"
	"errors"
	"gotest/internal/character"
)

var (
	ErrInvalidInput         = errors.New("world: invalid input")
	ErrCharacterRequired    = character.ErrSelectionRequired
	ErrCharacterChanged     = character.ErrSelectionChanged
	ErrPositionChanged      = errors.New("world: location changed")
	ErrNotFound             = errors.New("world: planet not found")
	ErrAccessDenied         = errors.New("world: travel not permitted")
	ErrLocationsUnavailable = errors.New("world: no locations available")
)

type Location struct {
	ID               int    `json:"id"`
	PlanetID         int    `json:"planet_id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Position         int    `json:"position"`
	RecommendedLevel int    `json:"recommended_level"`
}
type Position struct {
	CharacterID int      `json:"character_id"`
	Location    Location `json:"location"`
	Planet      Planet   `json:"planet"`
}
type Destination struct {
	PlanetID  int    `json:"planet_id"`
	CanTravel bool   `json:"can_travel"`
	Reason    string `json:"reason"`
}
type Navigation struct {
	Position       Position      `json:"position"`
	Destinations   []Destination `json:"destinations"`
	ReturnRequired bool          `json:"return_required"`
}
type TravelInput struct {
	CharacterID    int `json:"character_id"`
	PlanetID       int `json:"planet_id"`
	FromLocationID int `json:"from_location_id"`
}
type TravelResult struct {
	Position     Position `json:"position"`
	ReturnedHome bool     `json:"returned_home"`
}

// War permissions will be provided by the war module. Until then only current
// faction ownership permits visits; content readiness is an independent check.
func Access(factionID int, planet Planet, hasLocations bool) Destination {
	d := Destination{PlanetID: planet.ID, Reason: "invasion_required"}
	if planet.CoreFaction != nil {
		d.Reason = "raid_required"
	}
	if planet.ControllingFaction != nil && planet.ControllingFaction.ID == factionID {
		d.Reason = "locations_unavailable"
		if hasLocations {
			d.CanTravel = true
			d.Reason = "available"
		}
	}
	return d
}

func validID(id int) bool { return id > 0 && int64(id) <= 1<<31-1 }
func checkActor(ctx context.Context, a character.Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(a.UserID) || a.SessionHash == "" {
		return character.ErrUnauthenticated
	}
	return nil
}
func (s *Service) Navigation(ctx context.Context, a character.Actor, id int) (*Navigation, error) {
	if err := checkActor(ctx, a); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInvalidInput
	}
	return s.repository.Navigation(ctx, a, id)
}
func (s *Service) Travel(ctx context.Context, a character.Actor, input TravelInput) (*TravelResult, error) {
	if err := checkActor(ctx, a); err != nil {
		return nil, err
	}
	if !validID(input.CharacterID) || !validID(input.PlanetID) || !validID(input.FromLocationID) {
		return nil, ErrInvalidInput
	}
	return s.repository.Travel(ctx, a, input)
}
