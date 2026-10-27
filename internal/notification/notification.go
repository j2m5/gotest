// Package notification owns durable, encrypted account notifications.
package notification

import (
	"context"
	"errors"
	"time"
)

type Kind string

const (
	Verification     Kind = "verification"
	PasswordReset    Kind = "password_reset"
	PasswordChanged  Kind = "password_changed"
	ResetRequest     Kind = "reset_request"
	MaxAttempts           = 8
	LeaseDuration         = 120 * time.Second
	MaxResetRequests      = 1024
)

var (
	ErrLeaseLost = errors.New("notification lease is no longer owned")
	ErrQueueFull = errors.New("password reset request queue is full")
)

// Message carries only encrypted content across the persistence boundary.
type Message struct {
	ID         string
	Kind       Kind
	KeyID      string
	Ciphertext []byte `json:"-"`
	ExpiresAt  time.Time
}

type Content struct {
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}

type Job struct {
	Message
	Lease     string `json:"-"`
	Attempts  int
	CreatedAt time.Time
}

type Delivery struct {
	ID        string
	Kind      Kind
	Email     string `json:"-"`
	Token     string `json:"-"`
	CreatedAt time.Time
}

type Store interface {
	Claim(context.Context) (*Job, error)
	Owns(context.Context, Job) (bool, error)
	Complete(context.Context, Job) error
	Retry(context.Context, Job, time.Duration, bool) error
}

type Sender interface {
	Deliver(context.Context, Delivery) error
}

// ResetProcessor commits the generated token, outgoing mail and deletion of the
// incoming request together. The worker must not acknowledge it a second time.
type ResetProcessor func(context.Context, Job, Content) error
