// Package npc owns reusable enemy templates, placements and shared instances.
// Battle ownership, respawn and rewards will be connected by the combat module.
package npc

import (
	"context"
	"errors"

	"gotest/internal/character"
)

var ErrInvalidInput = errors.New("npc: invalid request")

type Instance struct {
	ID            int    `json:"id"`
	NPCTemplateID int    `json:"npc_template_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Kind          string `json:"kind"`
	Level         int    `json:"level"`
	Rank          string `json:"rank"`
}

type List struct {
	LocationID int        `json:"location_id"`
	NPCs       []Instance `json:"npcs"`
}

type Repository interface {
	List(context.Context, character.Actor, int, int) (*List, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, actor character.Actor, characterID, locationID int) (*List, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	validID := func(id int) bool { return id > 0 && int64(id) <= 1<<31-1 }
	if !validID(actor.UserID) || actor.SessionHash == "" {
		return nil, character.ErrUnauthenticated
	}
	if !validID(characterID) || !validID(locationID) {
		return nil, ErrInvalidInput
	}
	return s.repository.List(ctx, actor, characterID, locationID)
}
