package auth

import (
	"context"
	"gotest/internal/models"
	"time"
)

// Repository owns persistence and atomic state changes. Token arguments are SHA-256 digests.
type Repository interface {
	CreateRegistration(context.Context, string, string, string, string, time.Time) error
	FindUserByEmail(context.Context, string) (*models.User, error)
	CreateSession(context.Context, int, string, time.Time) error
	FindSession(context.Context, string) (*Session, error)
	// DeleteSession revokes the token and clears the account's shared character
	// selection atomically if the session still exists and has not expired.
	DeleteSession(context.Context, string) error
	ConsumeVerification(context.Context, string, time.Time) error
	ReplaceVerification(context.Context, int, string, time.Time) error
}
type VerificationSender interface {
	SendVerification(context.Context, string, string) error
}
type Session struct {
	User      *models.User
	ExpiresAt time.Time
}
type RegisterInput struct{ Email, Login, Password string }
type LoginResult struct {
	User      *models.User
	Token     string
	ExpiresAt time.Time
}
type Options struct{ SessionTTL, VerificationTTL time.Duration }
type Service struct {
	repository   Repository
	sender       VerificationSender
	options      Options
	now          func() time.Time
	passwordCost int
}
