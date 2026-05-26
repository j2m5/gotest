package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/flash"
)

type Handler struct {
	storage *db.Storage
	auth    *auth.Auth
	flash   *flash.Flash
}

func NewHandler(storage *db.Storage, auth *auth.Auth, flash *flash.Flash) *Handler {
	return &Handler{storage: storage, auth: auth, flash: flash}
}
