package routes_test

import (
	"context"

	"gotest/internal/character"
	"gotest/internal/presence"
)

type presenceStub struct {
	filled bool
	err    error
}

func (s *presenceStub) Heartbeat(context.Context, character.Actor, *int) error { return s.err }
func (s *presenceStub) Snapshot(_ context.Context, _ character.Actor, ids []int) (*presence.Snapshot, error) {
	result := &presence.Snapshot{Characters: []presence.Status{}}
	if s.filled {
		result.Total = 1
	}
	for _, id := range ids {
		result.Characters = append(result.Characters, presence.Status{CharacterID: id, IsOnline: s.filled && id == 10})
	}
	return result, s.err
}
func (s *presenceStub) Online(context.Context, character.Actor, int) (*presence.Directory, error) {
	result := &presence.Directory{Characters: []character.PublicProfile{}}
	if s.filled {
		result.Total = 51
		result.Characters = append(result.Characters, character.PublicProfile{ID: 10, Name: "Malgus", FactionID: 2, ClassID: 8, Level: 1, AccountRole: "moderator"})
		result.NextAfterID = new(10)
	}
	return result, s.err
}
