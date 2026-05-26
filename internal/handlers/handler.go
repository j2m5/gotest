package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/db"
)

type Handler struct {
	storage *db.Storage
	auth    *auth.Auth
}

func NewHandler(storage *db.Storage, auth *auth.Auth) *Handler {
	return &Handler{storage: storage, auth: auth}
}
