package auth

import (
	"gotest/internal/db"
	"gotest/internal/models"
	"net/http"
)

func CurrentUser(r *http.Request) *models.User {
	cookie, err := r.Cookie("session_token")

	if err != nil {
		return nil
	}

	user, err := db.FindUserBySessionToken(db.Pool, cookie.Value)

	if err != nil {
		return nil
	}

	return user
}
