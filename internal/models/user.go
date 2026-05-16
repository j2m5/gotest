package models

import "time"

type User struct {
	ID              int
	RoleID          int
	Name            *string
	Email           string
	Login           string
	Password        string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
