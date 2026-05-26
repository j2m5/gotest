package middleware

import (
	"gotest/internal/auth"
)

type Middleware struct {
	auth *auth.Auth
}

func NewMiddleware(auth *auth.Auth) *Middleware {
	return &Middleware{auth: auth}
}
