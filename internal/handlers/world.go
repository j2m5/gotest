package handlers

import (
	"context"
	"net/http"
	"strconv"

	"gotest/internal/character"
	"gotest/internal/httpapi"
	"gotest/internal/world"
)

type WorldService interface {
	Catalog(context.Context) (*world.Catalog, error)
	Navigation(context.Context, character.Actor, int) (*world.Navigation, error)
	Travel(context.Context, character.Actor, world.TravelInput) (*world.TravelResult, error)
	Locations(context.Context, character.Actor, int, int) (*world.LocationCatalog, error)
	Move(context.Context, character.Actor, world.MoveInput) (*world.TravelResult, error)
}

func (h *WorldHandler) Navigation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	if err != nil || id <= 0 {
		respondError(w, r, world.ErrInvalidInput)
		return
	}
	result, err := h.service.Navigation(r.Context(), characterActor(r.Context()), int(id))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
func (h *WorldHandler) Travel(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[world.TravelInput](w, r)
	if !ok {
		return
	}
	result, err := h.service.Travel(r.Context(), characterActor(r.Context()), *in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (h *WorldHandler) Locations(w http.ResponseWriter, r *http.Request) {
	characterID, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	planetID, planetErr := strconv.ParseInt(r.URL.Query().Get("planet_id"), 10, 32)
	if err != nil || planetErr != nil || characterID <= 0 || planetID <= 0 {
		respondError(w, r, world.ErrInvalidInput)
		return
	}
	result, err := h.service.Locations(r.Context(), characterActor(r.Context()), int(characterID), int(planetID))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
func (h *WorldHandler) Move(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[world.MoveInput](w, r)
	if !ok {
		return
	}
	result, err := h.service.Move(r.Context(), characterActor(r.Context()), *in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

type WorldHandler struct{ service WorldService }

func NewWorldHandler(service WorldService) *WorldHandler { return &WorldHandler{service: service} }

func (h *WorldHandler) Planets(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.Catalog(r.Context())
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, catalog)
}
