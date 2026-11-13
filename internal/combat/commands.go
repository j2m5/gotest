package combat

import "time"

type UseInput struct {
	CharacterID int    `json:"character_id"`
	AbilityID   int    `json:"ability_id"`
	Key         string `json:"request_key"`
}

type UseRequest struct {
	UseInput
	BattleID string `json:"battle_id"`
}
type RetreatInput struct {
	CharacterID int    `json:"character_id"`
	BattleID    string `json:"battle_id"`
}

type Result struct {
	State
	Replayed      bool `json:"replayed"`
	EventSequence int  `json:"event_sequence"`
}

type View struct {
	State
	NextActionAt *time.Time `json:"next_action_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	Events       []Event    `json:"events"`
	NextAfter    *int       `json:"next_after"`
	LogClearedAt *time.Time `json:"log_cleared_at"`
}

func ValidateUse(battleID int64, in UseInput) error {
	if battleID <= 0 || !ValidID(in.AbilityID) {
		return ErrInvalidInput
	}
	return ValidateStart(StartInput{CharacterID: in.CharacterID, LocationID: 1, NPCSpawnID: 1, Key: in.Key}, NPCStats{HP: 1})
}
