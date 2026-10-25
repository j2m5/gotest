package models

import "time"

type User struct {
	ID              int
	RoleID          int16
	CharacterID     *int // Shared active character; nil while choosing a character or after logout.
	Name            *string
	Email           string
	Login           string
	Password        string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
