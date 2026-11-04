package handlers

import (
	"context"
	"net/http"
	"strconv"

	"gotest/internal/character"
	"gotest/internal/httpapi"
	"gotest/internal/npc"
)

type NPCService interface {
	List(context.Context, character.Actor, int, int) (*npc.List, error)
}
type NPCHandler struct{ service NPCService }

func NewNPCHandler(service NPCService) *NPCHandler { return &NPCHandler{service: service} }

func (h *NPCHandler) List(w http.ResponseWriter, r *http.Request) {
	characterID, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	locationID, locationErr := strconv.ParseInt(r.URL.Query().Get("location_id"), 10, 32)
	if err != nil || locationErr != nil || characterID <= 0 || locationID <= 0 {
		respondError(w, r, npc.ErrInvalidInput)
		return
	}
	result, err := h.service.List(r.Context(), characterActor(r.Context()), int(characterID), int(locationID))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
