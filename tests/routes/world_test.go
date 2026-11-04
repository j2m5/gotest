package routes_test

import (
	"context"
	"gotest/internal/character"
	"gotest/internal/world"
)

type worldStub struct {
	empty bool
	err   error
}

func (s *worldStub) Navigation(_ context.Context, _ character.Actor, id int) (*world.Navigation, error) {
	return &world.Navigation{Position: s.position(id), Destinations: []world.Destination{{PlanetID: 1, CanTravel: true, Reason: "available"}, {PlanetID: 2, Reason: "invasion_required"}}, ReturnRequired: false}, s.err
}
func (s *worldStub) position(id int) world.Position {
	f := &world.Faction{ID: 103, Code: "galactic_republic", Name: "Республика"}
	return world.Position{CharacterID: id, Location: world.Location{ID: 7, PlanetID: 1, Code: "lower_levels", Name: "Нижние уровни", Description: "Район города", Position: 1, RecommendedLevel: 1}, Planet: world.Planet{ID: 1, Code: "coruscant", Name: "Корускант", NameEN: "Coruscant", CoreFaction: f, ControllingFaction: f}}
}
func (s *worldStub) Travel(_ context.Context, _ character.Actor, in world.TravelInput) (*world.TravelResult, error) {
	return &world.TravelResult{Position: s.position(in.CharacterID)}, s.err
}

func (s *worldStub) Locations(_ context.Context, _ character.Actor, id, _ int) (*world.LocationCatalog, error) {
	p := s.position(id)
	return &world.LocationCatalog{Position: p, Locations: []world.Location{p.Location}}, s.err
}
func (s *worldStub) Move(_ context.Context, _ character.Actor, in world.MoveInput) (*world.TravelResult, error) {
	return &world.TravelResult{Position: s.position(in.CharacterID)}, s.err
}

func (s *worldStub) Catalog(context.Context) (*world.Catalog, error) {
	result := &world.Catalog{Planets: []world.Planet{}}
	if !s.empty {
		faction := &world.Faction{ID: 103, Code: "galactic_republic", Name: "Республика"}
		result.Planets = []world.Planet{
			{ID: 1, Code: "coruscant", Name: "Корускант", NameEN: "Coruscant", CoreFaction: faction, ControllingFaction: faction},
			{ID: 2, Code: "tatooine", Name: "Татуин", NameEN: "Tatooine", Description: new("Пустынный мир")},
			{ID: 3, Code: "balmorra", Name: "Балморра", NameEN: "Balmorra", ControllingFaction: faction},
		}
	}
	return result, s.err
}
