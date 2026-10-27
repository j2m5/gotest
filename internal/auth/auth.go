package auth

import (
	"context"
	"time"
)

// Repository owns persistence and atomic state changes. Token arguments are SHA-256 digests.
type Repository interface {
	CreateRegistration(context.Context, Registration) error
	FindCredentialsByEmail(context.Context, string) (*Credentials, error)
	CreateSession(context.Context, SessionInput) error
	FindSession(context.Context, string) (*Session, error)
	// DeleteSession revokes the token and clears the account's shared character
	// selection atomically if the session still exists and has not expired.
	DeleteSession(context.Context, string) error
	ConsumeVerification(context.Context, string, time.Time) error
	ReplaceVerification(context.Context, int, string, time.Time) error
	CreatePasswordReset(context.Context, string, string, time.Time) (bool, error)
	ResetPassword(context.Context, string, string) (string, error)
}
type EmailSender interface {
	SendVerification(context.Context, string, string) error
	SendPasswordReset(context.Context, string, string) error
	SendPasswordChanged(context.Context, string) error
}

// Identity is the account data needed by authenticated application code. It
// deliberately cannot carry password hashes, session tokens or storage metadata.
type Identity struct {
	ID              int
	RoleID          int16
	CharacterID     *int
	Email           string
	Login           string
	EmailVerifiedAt *time.Time
}

// Credentials are loaded only when validating a password; they never become
// the current user or the value stored in the HTTP authentication context.
type Credentials struct {
	User         Identity
	PasswordHash string `json:"-"`
}

// Registration contains already normalized account data and one-way token and
// password hashes. The repository persists the account and verification atomically.
type Registration struct {
	Email                 string
	Login                 string
	PasswordHash          string `json:"-"`
	VerificationTokenHash string `json:"-"`
	VerificationExpiresAt time.Time
}

type Session struct {
	User      *Identity
	ExpiresAt time.Time
}

// SessionInput ties a login to the credentials that were actually verified.
// Persistence must serialize this comparison with password changes.
type SessionInput struct {
	UserID       int
	PasswordHash string `json:"-"`
	TokenHash    string `json:"-"`
	ExpiresAt    time.Time
}
type ResetPasswordInput struct{ Token, Password string }
type RegisterInput struct{ Email, Login, Password string }
type LoginInput struct {
	Email, Password string
	RememberMe      bool
}
type LoginResult struct {
	User       *Identity
	Token      string
	ExpiresAt  time.Time
	Persistent bool
}
type Options struct{ SessionTTL, RememberMeTTL, VerificationTTL, PasswordResetTTL time.Duration }
type Service struct {
	repository   Repository
	sender       EmailSender
	options      Options
	now          func() time.Time
	passwordCost int
}
