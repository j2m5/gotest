package models

import "time"

type Faction struct {
	ID          int
	Name        string
	Alias       string
	Description string
	Icon        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
