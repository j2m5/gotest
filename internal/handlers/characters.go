package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"gotest/internal/character"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
	"gotest/internal/validation"
)

type CharacterService interface {
	Catalog(context.Context) (*character.Catalog, error)
	List(context.Context, character.Actor) (*character.Roster, error)
	Create(context.Context, character.Actor, character.CreateInput) (*character.Character, error)
	SetActive(context.Context, character.Actor, *int) (*character.Character, error)
}

type CharacterHandler struct{ service CharacterService }

func NewCharacterHandler(service CharacterService) *CharacterHandler {
	return &CharacterHandler{service: service}
}

func (h *CharacterHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.Catalog(r.Context())
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, catalog)
}

func (h *CharacterHandler) List(w http.ResponseWriter, r *http.Request) {
	roster, err := h.service.List(r.Context(), characterActor(r.Context()))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, roster)
}

func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[character.CreateInput](w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(r.Context(), characterActor(r.Context()), *input)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"character": created})
}

func (h *CharacterHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[struct {
		CharacterID json.RawMessage `json:"character_id"`
	}](w, r)
	if !ok {
		return
	}
	// Missing and null are different operations: only an explicit null clears
	// selection. A malformed request must not affect other account sessions.
	var id *int
	if len(input.CharacterID) == 0 || json.Unmarshal(input.CharacterID, &id) != nil {
		message := "Укажите character_id: целое число или null"
		httpapi.WriteError(w, http.StatusBadRequest, httpapi.CodeInvalidRequest, message, validation.Fields{"character_id": {message}})
		return
	}
	selected, err := h.service.SetActive(r.Context(), characterActor(r.Context()), id)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"character": selected})
}

func characterActor(ctx context.Context) character.Actor {
	user := middleware.UserFromContext(ctx)
	if user == nil {
		return character.Actor{}
	}
	return character.Actor{UserID: user.ID, SessionHash: middleware.SessionHashFromContext(ctx)}
}
