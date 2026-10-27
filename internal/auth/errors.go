package auth

import "errors"

var (
	ErrNotFound           = errors.New("record not found")
	ErrConflict           = errors.New("email or login already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnverified         = errors.New("email not verified")
	ErrInvalidToken       = errors.New("invalid or expired verification token")
	ErrUnauthenticated    = errors.New("not authenticated")
	ErrAlreadyVerified    = errors.New("email already verified")
	ErrInvalidResetToken  = errors.New("invalid or expired password reset token")
	ErrResetUnavailable   = errors.New("password reset requests unavailable")
)
