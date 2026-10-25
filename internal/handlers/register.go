package handlers

import (
	"errors"
	"gotest/internal/auth"
	"log"
	"net/http"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeRequest[RegisterRequest](w, r)
	if !ok {
		return
	}
	err := h.auth.Register(r.Context(), auth.RegisterInput{Email: input.Email, Login: input.Login, Password: input.Password})
	if errors.Is(err, auth.ErrDelivery) {
		log.Printf("Registration saved; email delivery failed: %v", err)
		jsonResponse(w, map[string]any{"message": "Аккаунт создан, но письмо не отправлено. Повторите отправку на странице входа", "verification_sent": false}, 201)
		return
	}
	if err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, map[string]any{"message": "Аккаунт создан. Проверьте почту для подтверждения email", "verification_sent": true}, 201)
}
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || r.ContentLength != 0 {
		input, ok := decodeRequest[struct {
			Token string `json:"token"`
		}](w, r)
		if !ok {
			return
		}
		token = input.Token
	}
	if err := h.auth.VerifyEmail(r.Context(), token); err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, map[string]string{"message": "Email подтверждён. Теперь можно войти"}, 200)
}
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeRequest[LoginRequest](w, r)
	if !ok {
		return
	}
	if err := h.auth.ResendVerification(r.Context(), input.Email, input.Password); err != nil {
		h.respondError(w, r, err)
		return
	}
	jsonResponse(w, map[string]string{"message": "Письмо подтверждения отправлено"}, 200)
}
