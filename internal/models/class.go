package models

import "time"

type Class struct {
	ID          int
	Code        string
	FactionID   int
	Name        string
	Alias       *string
	Description *string
	Icon        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
