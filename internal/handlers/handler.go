package handlers

import (
	"encoding/json"
	"gotest/internal/auth"
	"gotest/internal/db"
	"log"
	"net/http"
)

type Handler struct {
	storage *db.Storage
	auth    *auth.Auth
}

func NewHandler(storage *db.Storage, auth *auth.Auth) *Handler {
	return &Handler{storage: storage, auth: auth}
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
