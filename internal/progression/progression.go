// Package progression applies server-authored experience rewards to characters.
package progression

import (
	"context"
	"errors"
)

var (
	ErrInvalidReward  = errors.New("invalid experience reward")
	ErrInvalidCurve   = errors.New("invalid experience curve")
	ErrInvalidState   = errors.New("invalid progression state")
	ErrNotFound       = errors.New("character not found")
	ErrRewardConflict = errors.New("experience reward key already used with a different amount")
)

type Level struct {
	Number             int16
	ExperienceRequired int64
}

type State struct {
	Level      int16
	Experience int64
}

type Change struct {
	Before, After State
	Granted       int64
}

// Reward must be constructed by a trusted server operation after checking that
// the character earned it. Key identifies one source event for this character.
type Reward struct {
	CharacterID int
	Key         string
	Experience  int64
}

// Result describes the original award, including on replay. Fetch the character
// separately if its current state is needed after other awards have been made.
type Result struct {
	CharacterID int
	Key         string
	Requested   int64
	Change      Change
	Replayed    bool
}

type Repository interface {
	Grant(context.Context, Reward) (*Result, error)
}
