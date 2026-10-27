package handlers

import (
	"errors"
	"net/http"

	"gotest/internal/auth"
	"gotest/internal/httpapi"
	"gotest/internal/validation"
)

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[struct {
		Email string `json:"email"`
	}](w, r)
	if !ok {
		return
	}
	err := h.resetRequests.RequestPasswordReset(r.Context(), input.Email)
	var invalid *validation.Error
	if errors.As(err, &invalid) {
		respondError(w, r, err)
		return
	}
	if err != nil {
		httpapi.WriteError(w, 503, "password_reset_unavailable", "Запросы восстановления временно недоступны. Повторите позже", nil)
		return
	}
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Если аккаунт с такой почтой существует, мы отправим ссылку для смены пароля. Повторный запрос возможен через минуту."})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}](w, r)
	if !ok {
		return
	}
	err := h.auth.ResetPassword(r.Context(), auth.ResetPasswordInput{Token: input.Token, Password: input.Password})
	if errors.Is(err, auth.ErrInvalidResetToken) {
		httpapi.WriteError(w, 400, "password_reset_token_invalid", "Ссылка смены пароля недействительна или истекла. Запросите новую", nil)
		return
	} else if err != nil {
		respondError(w, r, err)
		return
	}
	h.clearSessionCookie(w)
	httpapi.WriteJSON(w, 200, map[string]string{"message": "Пароль изменён. Войдите с новым паролем. Все прежние сессии завершены."})
}
