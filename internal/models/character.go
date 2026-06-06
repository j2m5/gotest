package models

import "time"

type Character struct {
	ID          int
	UserID      int
	FactionID   int
	ClassID     int
	Name        string
	Role        string
	Level       int16
	Experience  int
	Avatar      string
	Hp          int
	Mp          int
	HpCurrent   int
	MpCurrent   int
	Mastery     int
	Endurance   int
	Power       int
	Accuracy    int
	Alacrity    int
	Critical    int
	Defense     int
	Shield      int
	Absorption  int
	ActivatedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
