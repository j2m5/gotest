// Package combat owns the persistent lifecycle of shared-NPC encounters.
package combat

import (
	"errors"
	"math"
	"time"

	"gotest/internal/character"
)

var (
	ErrInvalidInput   = errors.New("combat: invalid request")
	ErrInBattle       = errors.New("combat: character already in battle")
	ErrNPCUnavailable = errors.New("combat: NPC unavailable")
	ErrConflict       = errors.New("combat: start key belongs to another request")
	ErrNotFound       = errors.New("combat: battle not found")
	ErrDefeated       = errors.New("combat: character has no health")
	ErrInvalidRules   = errors.New("combat: invalid server rules")
	ErrInvalidState   = errors.New("combat: invalid execution state")
	ErrFinished       = errors.New("combat: battle finished")
)

type Status string

const (
	Active    Status = "active"
	Retreated Status = "retreated"
	Defeated  Status = "defeated"
	Won       Status = "won"
	Expired   Status = "expired"
)

type Opponent struct {
	SpawnID    int    `json:"npc_spawn_id"`
	TemplateID int    `json:"npc_template_id"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Rank       string `json:"rank"`
	HP         int    `json:"hp"`
	HPCurrent  int    `json:"hp_current"`
}

type Battle struct {
	ID          int64      `json:"id,string"`
	CharacterID int        `json:"character_id"`
	LocationID  int        `json:"location_id"`
	Status      Status     `json:"status"`
	NPC         Opponent   `json:"npc"`
	CreatedAt   time.Time  `json:"created_at"`
	EndedAt     *time.Time `json:"ended_at"`
}

type State struct {
	Character *character.Character `json:"character"`
	Battle    *Battle              `json:"battle"`
}

// StartInput identifies an action. NPCStats is deliberately separate: combat
// balance is supplied by trusted server code, never by a player's request.
type StartInput struct {
	CharacterID int    `json:"character_id"`
	LocationID  int    `json:"location_id"`
	NPCSpawnID  int    `json:"npc_spawn_id"`
	Key         string `json:"request_key"`
}
type NPCStats struct{ HP int }
type Change struct {
	Battle   *Battle
	Replayed bool
}

func ValidID(id int) bool { return id > 0 && int64(id) <= math.MaxInt32 }

func ValidateStart(in StartInput, stats NPCStats) error {
	if !ValidID(in.CharacterID) || !ValidID(in.LocationID) || !ValidID(in.NPCSpawnID) || !ValidID(stats.HP) || len(in.Key) == 0 || len(in.Key) > 128 {
		return ErrInvalidInput
	}
	for i := 0; i < len(in.Key); i++ {
		b := in.Key[i]
		alphanumeric := b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
		if !alphanumeric && (i == 0 || b != ':' && b != '.' && b != '_' && b != '/' && b != '-') {
			return ErrInvalidInput
		}
	}
	return nil
}
