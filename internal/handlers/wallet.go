package handlers

import (
	"context"
	"net/http"
	"strconv"

	"gotest/internal/character"
	"gotest/internal/httpapi"
	"gotest/internal/wallet"
)

type WalletService interface {
	Balances(context.Context, character.Actor, int) (*wallet.Balances, error)
	History(context.Context, character.Actor, int, int64) (*wallet.History, error)
}
type WalletHandler struct{ service WalletService }

func NewWalletHandler(service WalletService) *WalletHandler { return &WalletHandler{service: service} }

func walletCharacterID(r *http.Request) (int, error) {
	id, err := strconv.ParseInt(r.URL.Query().Get("character_id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, wallet.ErrInvalidInput
	}
	return int(id), nil
}
func (h *WalletHandler) Balances(w http.ResponseWriter, r *http.Request) {
	id, err := walletCharacterID(r)
	if err != nil {
		respondError(w, r, err)
		return
	}
	result, err := h.service.Balances(r.Context(), characterActor(r.Context()), id)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
func (h *WalletHandler) History(w http.ResponseWriter, r *http.Request) {
	id, err := walletCharacterID(r)
	if err != nil {
		respondError(w, r, err)
		return
	}
	var before int64
	if values, exists := r.URL.Query()["before"]; exists {
		if len(values) != 1 || len(values[0]) == 0 || values[0][0] < '1' || values[0][0] > '9' {
			respondError(w, r, wallet.ErrInvalidInput)
			return
		}
		for _, b := range values[0] {
			if b < '0' || b > '9' {
				respondError(w, r, wallet.ErrInvalidInput)
				return
			}
		}
		before, err = strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			respondError(w, r, wallet.ErrInvalidInput)
			return
		}
	}
	result, err := h.service.History(r.Context(), characterActor(r.Context()), id, before)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}
