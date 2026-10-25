package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gotest/internal/character"
	"gotest/internal/middleware"
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
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, catalog, http.StatusOK)
}

func (h *CharacterHandler) List(w http.ResponseWriter, r *http.Request) {
	roster, err := h.service.List(r.Context(), characterActor(r.Context()))
	if err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, roster, http.StatusOK)
}

func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeRequest[character.CreateInput](w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(r.Context(), characterActor(r.Context()), *input)
	if err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, map[string]any{"character": created}, http.StatusCreated)
}

func (h *CharacterHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeRequest[struct {
		CharacterID json.RawMessage `json:"character_id"`
	}](w, r)
	if !ok {
		return
	}
	// Missing and null are different operations: only an explicit null clears
	// selection. A malformed request must not affect other account sessions.
	var id *int
	if len(input.CharacterID) == 0 || json.Unmarshal(input.CharacterID, &id) != nil {
		jsonError(w, "Укажите character_id: целое число или null", http.StatusBadRequest)
		return
	}
	selected, err := h.service.SetActive(r.Context(), characterActor(r.Context()), id)
	if err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, map[string]any{"character": selected}, http.StatusOK)
}

func characterActor(ctx context.Context) character.Actor {
	user := middleware.UserFromContext(ctx)
	if user == nil {
		return character.Actor{}
	}
	return character.Actor{UserID: user.ID, SessionHash: middleware.SessionHashFromContext(ctx)}
}

func (h *CharacterHandler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *character.ValidationError
	switch {
	case errors.As(err, &validation):
		jsonResponse(w, map[string]any{"error": "Проверьте заполненные поля", "errors": validation.Messages}, http.StatusUnprocessableEntity)
	case errors.Is(err, character.ErrUnauthenticated):
		jsonError(w, "Необходимо войти в аккаунт", http.StatusUnauthorized)
	case errors.Is(err, character.ErrNotFound):
		jsonError(w, "Персонаж не найден", http.StatusNotFound)
	case errors.Is(err, character.ErrNameTaken):
		jsonError(w, "Это имя персонажа уже занято", http.StatusConflict)
	case errors.Is(err, character.ErrInvalidClass):
		jsonError(w, "Выбранный класс недоступен для этой фракции", http.StatusUnprocessableEntity)
	default:
		log.Printf("ERROR: %s %s: %v", r.Method, r.URL.Path, err)
		jsonError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}
