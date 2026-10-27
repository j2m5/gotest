package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/httpapi"
	"gotest/internal/observability"
	"gotest/internal/validation"
)

// Domain errors are translated at the HTTP boundary. Modules do not depend on
// transport codes, and wrapped errors keep their original meaning.
func respondError(w http.ResponseWriter, r *http.Request, err error) {
	var invalid *validation.Error
	switch {
	case errors.As(err, &invalid):
		httpapi.WriteError(w, 422, httpapi.CodeValidationFailed, "Проверьте заполненные поля", invalid.Fields)
	case errors.Is(err, auth.ErrUnauthenticated), errors.Is(err, character.ErrUnauthenticated):
		httpapi.WriteError(w, 401, httpapi.CodeUnauthenticated, "Необходимо войти в аккаунт", nil)
	case errors.Is(err, auth.ErrInvalidCredentials):
		httpapi.WriteError(w, 401, "invalid_credentials", "Неверный email или пароль", nil)
	case errors.Is(err, auth.ErrUnverified):
		httpapi.WriteError(w, 403, "email_unverified", "Подтвердите email перед входом", nil)
	case errors.Is(err, auth.ErrConflict):
		httpapi.WriteError(w, 409, "account_conflict", "Email или логин уже используется", nil)
	case errors.Is(err, auth.ErrAlreadyVerified):
		httpapi.WriteError(w, 409, "email_already_verified", "Email уже подтверждён", nil)
	case errors.Is(err, auth.ErrInvalidToken):
		httpapi.WriteError(w, 400, "verification_token_invalid", "Ссылка подтверждения недействительна или истекла", nil)
	case errors.Is(err, character.ErrNotFound):
		httpapi.WriteError(w, 404, "character_not_found", "Персонаж не найден", nil)
	case errors.Is(err, character.ErrNameTaken):
		message := "Это имя персонажа уже занято"
		httpapi.WriteError(w, 409, "character_name_taken", message, validation.Fields{"name": {message}})
	case errors.Is(err, character.ErrInvalidClass):
		message := "Выбранный класс недоступен для этой фракции"
		httpapi.WriteError(w, 422, "character_class_mismatch", message, validation.Fields{"class_id": {message}})
	default:
		observability.Logger(r.Context()).LogAttrs(r.Context(), slog.LevelError, "request_failed", observability.ErrorAttrs(err)...)
		httpapi.WriteError(w, 500, httpapi.CodeInternalError, "Внутренняя ошибка сервера", nil)
	}
}
