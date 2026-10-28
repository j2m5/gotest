package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"gotest/internal/httpapi"
	"gotest/internal/middleware"
	"gotest/internal/social"
)

type SocialHandler struct {
	service social.Repository
	notify  func()
}

func NewSocialHandler(service social.Repository, notify func()) *SocialHandler {
	if notify == nil {
		notify = func() {}
	}
	return &SocialHandler{service, notify}
}

func (h *SocialHandler) Moderators(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.Moderators(r.Context(), socialActor(r.Context()))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, list)
}

func socialActor(ctx context.Context) social.Actor {
	user := middleware.UserFromContext(ctx)
	if user == nil {
		return social.Actor{}
	}
	return social.Actor{UserID: user.ID, SessionHash: middleware.SessionHashFromContext(ctx)}
}

func queryID(w http.ResponseWriter, r *http.Request, key string, bits int) (int64, bool) {
	values, present := r.URL.Query()[key]
	if !present {
		return 0, true
	}
	if len(values) == 1 && values[0] != "" {
		digits := true
		for _, c := range values[0] {
			if c < '0' || c > '9' {
				digits = false
				break
			}
		}
		value, err := strconv.ParseInt(values[0], 10, bits)
		if digits && err == nil && value > 0 {
			return value, true
		}
	}
	httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Некорректный параметр "+key, nil)
	return 0, false
}

func (h *SocialHandler) Chat(w http.ResponseWriter, r *http.Request) {
	before, ok := queryID(w, r, "before_id", 64)
	if !ok {
		return
	}
	page, err := h.service.Chat(r.Context(), socialActor(r.Context()), before)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, page)
}

func (h *SocialHandler) Send(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[social.SendInput](w, r)
	if !ok {
		return
	}
	message, err := h.service.Send(r.Context(), socialActor(r.Context()), *in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	h.notify()
	httpapi.WriteJSON(w, 201, map[string]any{"message": message})
}

func (h *SocialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[social.DeleteInput](w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), socialActor(r.Context()), *in); err != nil {
		respondError(w, r, err)
		return
	}
	h.notify()
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Сообщение удалено"})
}

func (h *SocialHandler) Bans(w http.ResponseWriter, r *http.Request) {
	before, ok := queryID(w, r, "before_id", 64)
	if !ok {
		return
	}
	user, ok := queryID(w, r, "user_id", 32)
	if !ok {
		return
	}
	page, err := h.service.Bans(r.Context(), socialActor(r.Context()), before, int(user))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, page)
}

func (h *SocialHandler) Ban(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[struct {
		UserID   int             `json:"user_id"`
		Scope    string          `json:"scope"`
		Reason   string          `json:"reason"`
		Duration json.RawMessage `json:"duration_hours"`
	}](w, r)
	if !ok {
		return
	}
	// An omitted duration must never accidentally turn into a permanent ban.
	var hours *int
	if len(in.Duration) == 0 || json.Unmarshal(in.Duration, &hours) != nil {
		httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Укажите duration_hours: число часов или null для бессрочного бана аккаунта", nil)
		return
	}
	ban, err := h.service.Ban(r.Context(), socialActor(r.Context()), social.BanInput{UserID: in.UserID, Scope: in.Scope, Reason: in.Reason, DurationHours: hours})
	if err != nil {
		respondError(w, r, err)
		return
	}
	h.notify()
	httpapi.WriteJSON(w, 201, map[string]any{"ban": ban})
}

func (h *SocialHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSON[social.RevokeInput](w, r)
	if !ok {
		return
	}
	if err := h.service.Revoke(r.Context(), socialActor(r.Context()), *in); err != nil {
		respondError(w, r, err)
		return
	}
	h.notify()
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Бан снят"})
}

func (h *SocialHandler) Actions(w http.ResponseWriter, r *http.Request) {
	before, ok := queryID(w, r, "before_id", 64)
	if !ok {
		return
	}
	page, err := h.service.Actions(r.Context(), socialActor(r.Context()), before)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, page)
}
