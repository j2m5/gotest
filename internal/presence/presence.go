// Package presence tracks recent player activity, independently of session lifetime.
package presence

import (
	"context"
	"errors"
	"time"

	"gotest/internal/character"
)

const (
	OnlineWindow  = 2 * time.Minute
	WriteInterval = 15 * time.Second
	MaxStatusIDs  = 100
	PageSize      = 50
)

var (
	ErrInvalidInput     = errors.New("presence: invalid input")
	ErrCharacterChanged = errors.New("presence: active character changed")
)

type Status struct {
	CharacterID int  `json:"character_id"`
	IsOnline    bool `json:"is_online"`
}

type Snapshot struct {
	Total      int      `json:"total"`
	Characters []Status `json:"characters"`
}

type Directory struct {
	Total       int                       `json:"total"`
	Characters  []character.PublicProfile `json:"characters"`
	NextAfterID *int                      `json:"next_after_id"`
}

type Repository interface {
	Heartbeat(context.Context, character.Actor, *int) error
	Snapshot(context.Context, character.Actor, []int) (*Snapshot, error)
	Online(context.Context, character.Actor, int) (*Directory, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func checkActor(ctx context.Context, actor character.Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if actor.UserID <= 0 || actor.SessionHash == "" {
		return character.ErrUnauthenticated
	}
	return nil
}

func (s *Service) Heartbeat(ctx context.Context, actor character.Actor, id *int) error {
	if err := checkActor(ctx, actor); err != nil {
		return err
	}
	if id != nil && (*id <= 0 || int64(*id) > 1<<31-1) {
		return ErrInvalidInput
	}
	return s.repository.Heartbeat(ctx, actor, id)
}

func (s *Service) Snapshot(ctx context.Context, actor character.Actor, ids []int) (*Snapshot, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	if len(ids) > MaxStatusIDs {
		return nil, ErrInvalidInput
	}
	seen := make(map[int]bool, len(ids))
	unique := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || int64(id) > 1<<31-1 {
			return nil, ErrInvalidInput
		}
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	return s.repository.Snapshot(ctx, actor, unique)
}

func (s *Service) Online(ctx context.Context, actor character.Actor, after int) (*Directory, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	if after < 0 || int64(after) > 1<<31-1 {
		return nil, ErrInvalidInput
	}
	return s.repository.Online(ctx, actor, after)
}
