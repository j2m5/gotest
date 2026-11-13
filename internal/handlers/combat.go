package handlers

import (
	"context"
	"net/http"
	"strconv"

	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/httpapi"
)

type CombatService interface {
	Current(context.Context, character.Actor, int) (*combat.State, error)
	Start(context.Context, character.Actor, combat.StartInput) (*combat.Result, error)
	Use(context.Context, character.Actor, int64, combat.UseInput) (*combat.Result, error)
	Retreat(context.Context, character.Actor, int64, int) (*combat.Result, error)
	View(context.Context, character.Actor, int64, int, int) (*combat.View, error)
}

type CombatHandler struct{ service CombatService }

func NewCombatHandler(service CombatService) *CombatHandler { return &CombatHandler{service: service} }

func (h *CombatHandler) Current(w http.ResponseWriter, r *http.Request) {
	characterID, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	if err != nil || characterID <= 0 {
		respondError(w, r, combat.ErrInvalidInput)
		return
	}
	state, err := h.service.Current(r.Context(), characterActor(r.Context()), int(characterID))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, state)
}

func parseBattleID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != value {
		return 0, combat.ErrInvalidInput
	}
	return id, nil
}

func (h *CombatHandler) Start(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[combat.StartInput](w, r)
	if !ok {
		return
	}
	result, err := h.service.Start(r.Context(), characterActor(r.Context()), *in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}
func (h *CombatHandler) Use(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[combat.UseRequest](w, r)
	if !ok {
		return
	}
	id, err := parseBattleID(in.BattleID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	result, err := h.service.Use(r.Context(), characterActor(r.Context()), id, in.UseInput)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}
func (h *CombatHandler) Retreat(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[combat.RetreatInput](w, r)
	if !ok {
		return
	}
	id, err := parseBattleID(in.BattleID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	result, err := h.service.Retreat(r.Context(), characterActor(r.Context()), id, in.CharacterID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}
func (h *CombatHandler) View(w http.ResponseWriter, r *http.Request) {
	id, err := parseBattleID(r.URL.Query().Get("battle_id"))
	characterID, charErr := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	var after int64
	var afterErr error
	if r.URL.Query().Has("after_sequence") {
		after, afterErr = strconv.ParseInt(r.URL.Query().Get("after_sequence"), 10, 32)
	}
	if err != nil || charErr != nil || characterID <= 0 || afterErr != nil || after < 0 {
		respondError(w, r, combat.ErrInvalidInput)
		return
	}
	result, err := h.service.View(r.Context(), characterActor(r.Context()), id, int(characterID), int(after))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}
