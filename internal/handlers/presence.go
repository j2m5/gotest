package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gotest/internal/httpapi"
	"gotest/internal/presence"
)

type PresenceHandler struct{ service *presence.Service }

func NewPresenceHandler(service *presence.Service) *PresenceHandler {
	return &PresenceHandler{service: service}
}

func (h *PresenceHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[struct {
		CharacterID json.RawMessage `json:"character_id"`
	}](w, r)
	if !ok {
		return
	}
	var id *int
	if len(in.CharacterID) == 0 || json.Unmarshal(in.CharacterID, &id) != nil {
		respondError(w, r, presence.ErrInvalidInput)
		return
	}
	if err := h.service.Heartbeat(r.Context(), characterActor(r.Context()), id); err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, struct {
		OK bool `json:"ok"`
	}{true})
}

func (h *PresenceHandler) Snapshot(w http.ResponseWriter, r *http.Request) {
	ids := []int{}
	values := r.URL.Query()["character_ids"]
	if len(values) > 1 {
		respondError(w, r, presence.ErrInvalidInput)
		return
	}
	if len(values) == 1 && values[0] != "" {
		parts := strings.Split(values[0], ",")
		if len(parts) > presence.MaxStatusIDs {
			respondError(w, r, presence.ErrInvalidInput)
			return
		}
		for _, part := range parts {
			id, err := strconv.ParseInt(part, 10, 32)
			if err != nil || id <= 0 {
				respondError(w, r, presence.ErrInvalidInput)
				return
			}
			ids = append(ids, int(id))
		}
	}
	result, err := h.service.Snapshot(r.Context(), characterActor(r.Context()), ids)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}

func (h *PresenceHandler) Online(w http.ResponseWriter, r *http.Request) {
	var after int64
	values := r.URL.Query()["after_id"]
	if len(values) > 1 {
		respondError(w, r, presence.ErrInvalidInput)
		return
	}
	if len(values) == 1 {
		var err error
		after, err = strconv.ParseInt(values[0], 10, 32)
		if err != nil || after < 0 {
			respondError(w, r, presence.ErrInvalidInput)
			return
		}
	}
	result, err := h.service.Online(r.Context(), characterActor(r.Context()), int(after))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, result)
}
