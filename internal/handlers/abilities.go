package handlers

import (
	"context"
	"net/http"
	"strconv"

	"gotest/internal/ability"
	"gotest/internal/character"
	"gotest/internal/httpapi"
)

type AbilityService interface {
	List(context.Context, character.Actor, int) (*ability.Loadout, error)
}

type AbilityHandler struct{ service AbilityService }

func NewAbilityHandler(service AbilityService) *AbilityHandler {
	return &AbilityHandler{service: service}
}

func (h *AbilityHandler) List(w http.ResponseWriter, r *http.Request) {
	characterID, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	if err != nil || characterID <= 0 {
		respondError(w, r, ability.ErrInvalidInput)
		return
	}
	result, err := h.service.List(r.Context(), characterActor(r.Context()), int(characterID))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
