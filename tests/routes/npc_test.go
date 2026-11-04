package routes_test

import (
	"context"

	"gotest/internal/character"
	"gotest/internal/npc"
)

type npcStub struct {
	empty bool
	err   error
}

func (s *npcStub) List(_ context.Context, _ character.Actor, _, locationID int) (*npc.List, error) {
	result := &npc.List{LocationID: locationID, NPCs: []npc.Instance{}}
	if !s.empty {
		result.NPCs = append(result.NPCs, npc.Instance{ID: 12, NPCTemplateID: 4, Name: "Дроид-страж", Description: "Охраняет проход", Kind: "droid", Level: 5, Rank: "elite"})
	}
	return result, s.err
}
