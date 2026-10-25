package middleware

import (
	"context"
	"gotest/internal/auth"
)

const SessionCookie = "session_token"

type Authenticator interface {
	CurrentUser(context.Context, string) (*auth.Identity, error)
}
type Middleware struct{ auth Authenticator }

func NewMiddleware(service Authenticator) *Middleware { return &Middleware{auth: service} }

type userContextKey struct{}
type sessionHashContextKey struct{}

func UserFromContext(ctx context.Context) *auth.Identity {
	user, _ := ctx.Value(userContextKey{}).(*auth.Identity)
	return user
}

// SessionHashFromContext returns the storage key of the authenticated session.
// It is used to recheck revocation inside account mutation transactions.
func SessionHashFromContext(ctx context.Context) string {
	hash, _ := ctx.Value(sessionHashContextKey{}).(string)
	return hash
}
