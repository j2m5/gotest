package handlers

import (
	"encoding/json"
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

func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, map[string]string{"error": message}, status)
}

func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR: %s %s - %v", r.Method, r.URL, err)
	jsonError(w, err.Error(), http.StatusInternalServerError)
}
