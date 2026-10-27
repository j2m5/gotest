package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/httpapi"
	"net/http"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[RegisterRequest](w, r)
	if !ok {
		return
	}
	err := h.auth.Register(r.Context(), auth.RegisterInput{Email: input.Email, Login: input.Login, Password: input.Password})
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 201, map[string]any{"message": "Аккаунт создан. Письмо подтверждения поставлено в очередь", "verification_queued": true})
}
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || r.ContentLength != 0 {
		input, ok := httpapi.DecodeJSON[struct {
			Token string `json:"token"`
		}](w, r)
		if !ok {
			return
		}
		token = input.Token
	}
	if err := h.auth.VerifyEmail(r.Context(), token); err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Email подтверждён. Теперь можно войти"})
}
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[LoginRequest](w, r)
	if !ok {
		return
	}
	if err := h.auth.ResendVerification(r.Context(), input.Email, input.Password); err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Письмо подтверждения поставлено в очередь"})
}
