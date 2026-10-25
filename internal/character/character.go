// Package character owns account character creation and selection rules.
package character

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrUnauthenticated = errors.New("character: session is not active")
	ErrNotFound        = errors.New("character: not found")
	ErrNameTaken       = errors.New("character: name is already in use")
	ErrInvalidClass    = errors.New("character: class does not belong to faction")
)

type ValidationError struct{ Messages []string }

func (e *ValidationError) Error() string { return strings.Join(e.Messages, "; ") }

// Actor identifies the authenticated caller. SessionHash is a storage key, never
// a client-supplied account ID or an unhashed cookie. Repositories must recheck
// the session after locking the account before making any changes.
type Actor struct {
	UserID      int
	SessionHash string
}

type CreateInput struct {
	Name      string `json:"name"`
	FactionID int    `json:"faction_id"`
	ClassID   int    `json:"class_id"`
}

// Character is the account-facing character view; storage and owner details
// deliberately remain in the persistence layer.
type Character struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	FactionID  int     `json:"faction_id"`
	ClassID    int     `json:"class_id"`
	Role       string  `json:"role"`
	Level      int16   `json:"level"`
	Experience int     `json:"experience"`
	Avatar     *string `json:"avatar"`
	HP         int     `json:"hp"`
	MP         int     `json:"mp"`
	HPCurrent  int     `json:"hp_current"`
	MPCurrent  int     `json:"mp_current"`
}

type Faction struct {
	ID          int     `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type Class struct {
	ID          int     `json:"id"`
	FactionID   int     `json:"faction_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type Catalog struct {
	Factions []Faction `json:"factions"`
	Classes  []Class   `json:"classes"`
}

type Roster struct {
	Characters        []Character `json:"characters"`
	ActiveCharacterID *int        `json:"active_character_id"`
}

type Repository interface {
	Catalog(context.Context) (*Catalog, error)
	List(context.Context, Actor) (*Roster, error)
	Create(context.Context, Actor, CreateInput) (*Character, error)
	// A nil ID clears the shared selection and returns a nil character.
	SetActive(context.Context, Actor, *int) (*Character, error)
}
