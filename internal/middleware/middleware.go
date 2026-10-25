package middleware

import (
	"context"
	"gotest/internal/models"
)

const SessionCookie = "session_token"

type Authenticator interface {
	CurrentUser(context.Context, string) (*models.User, error)
}
type Middleware struct{ auth Authenticator }

func NewMiddleware(service Authenticator) *Middleware { return &Middleware{auth: service} }

type userContextKey struct{}

func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userContextKey{}).(*models.User)
	return user
}
