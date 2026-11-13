// Package ability defines learned abilities and server-authoritative readiness.
package ability

import (
	"errors"
	"time"

	"gotest/internal/character"
)

const GlobalCooldownDuration = 1500 * time.Millisecond

var (
	ErrInvalidInput   = errors.New("ability: invalid request")
	ErrUnavailable    = errors.New("ability: not learned or unavailable to class")
	ErrCooldown       = errors.New("ability: on cooldown")
	ErrGlobalCooldown = errors.New("ability: global cooldown")
	ErrDefeated       = errors.New("ability: character has no health")
)

type Availability string

const (
	Ready              Availability = "ready"
	Cooldown           Availability = "cooldown"
	GlobalCooldown     Availability = "global_cooldown"
	InsufficientEnergy Availability = "insufficient_energy"
	Defeated           Availability = "defeated"
)

// Damage and Heal are base effect amounts, not percentages of weapon damage.
// The combat module applies the eventual attacker/target formulas.
type Definition struct {
	ID          int     `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	Damage      int     `json:"damage"`
	Heal        int     `json:"heal"`
	MPCost      int     `json:"mp_cost"`
	CooldownMS  int     `json:"cooldown_ms"`
}

type Owned struct {
	Definition
	ReadyAt      *time.Time   `json:"ready_at"`
	Availability Availability `json:"availability"`
}

type Loadout struct {
	Character        *character.Character `json:"character"`
	ObservedAt       time.Time            `json:"observed_at"`
	GlobalReadyAt    *time.Time           `json:"global_ready_at"`
	GlobalCooldownMS int64                `json:"global_cooldown_ms"`
	Abilities        []Owned              `json:"abilities"`
}

// Use describes a preparation inside the caller's uncommitted combat action.
type Use struct {
	Ability                        Definition
	UsedAt, ReadyAt, GlobalReadyAt time.Time
	EnergyRemaining                int
}

func AvailabilityAt(definition Definition, readyAt, globalReadyAt *time.Time, hp, mp int, at time.Time) Availability {
	if hp <= 0 {
		return Defeated
	}
	if readyAt != nil && at.Before(*readyAt) {
		return Cooldown
	}
	if globalReadyAt != nil && at.Before(*globalReadyAt) {
		return GlobalCooldown
	}
	if mp < definition.MPCost {
		return InsufficientEnergy
	}
	return Ready
}

func ReadinessError(state Availability) error {
	switch state {
	case Ready:
		return nil
	case Cooldown:
		return ErrCooldown
	case GlobalCooldown:
		return ErrGlobalCooldown
	case InsufficientEnergy:
		return character.ErrInsufficientEnergy
	case Defeated:
		return ErrDefeated
	default:
		return ErrUnavailable
	}
}
