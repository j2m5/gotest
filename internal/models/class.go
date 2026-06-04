package models

import "time"

type Class struct {
	ID          int
	FactionID   int
	Name        string
	Alias       string
	Description string
	Icon        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
