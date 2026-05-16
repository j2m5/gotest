package models

import "time"

type EmailVerification struct {
	ID        int
	UserID    int
	Token     string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
