package auth

import (
	"gotest/internal/db"
	"gotest/internal/models"
	"net/http"
)

type Auth struct {
	storage *db.Storage
}

func NewAuth(storage *db.Storage) *Auth {
	return &Auth{storage: storage}
}

func (a *Auth) CurrentUser(r *http.Request) *models.User {
	cookie, err := r.Cookie("session_token")

	if err != nil {
		return nil
	}

	user, err := a.storage.FindUserBySessionToken(cookie.Value)

	if err != nil {
		return nil
	}

	return user
}
