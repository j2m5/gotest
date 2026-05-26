package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/flash"
	"log"
	"net/http"
)

type Handler struct {
	storage *db.Storage
	auth    *auth.Auth
	flash   *flash.Flash
}

func NewHandler(storage *db.Storage, auth *auth.Auth, flash *flash.Flash) *Handler {
	return &Handler{storage: storage, auth: auth, flash: flash}
}

func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR: %s %s - %v", r.Method, r.URL, err)
	h.flash.Set(w, r, "error", "Произошла ошибка, повторите позже")
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
