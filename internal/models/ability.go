package models

import "time"

type Ability struct {
	ID          int
	Name        string
	Description string
	Icon        string
	Damage      int
	Heal        int
	MpCost      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
