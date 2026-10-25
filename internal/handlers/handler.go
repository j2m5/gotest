package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/models"
	"io"
	"log"
	"mime"
	"net/http"
)

type AuthService interface {
	Register(context.Context, auth.RegisterInput) error
	Login(context.Context, string, string) (*auth.LoginResult, error)
	Logout(context.Context, string) error
	VerifyEmail(context.Context, string) error
	ResendVerification(context.Context, string, string) error
	CurrentUser(context.Context, string) (*models.User, error)
}
type Handler struct {
	auth          AuthService
	secureCookies bool
}

func NewHandler(service AuthService, secureCookies bool) *Handler {
	return &Handler{auth: service, secureCookies: secureCookies}
}
func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func jsonError(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, map[string]string{"error": message}, status)
}
func (h *Handler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *auth.ValidationError
	switch {
	case errors.As(err, &validation):
		jsonResponse(w, map[string]any{"error": "Проверьте заполненные поля", "errors": validation.Messages}, 422)
	case errors.Is(err, auth.ErrInvalidCredentials):
		jsonError(w, "Неверный email или пароль", 401)
	case errors.Is(err, auth.ErrUnauthenticated):
		jsonError(w, "Необходимо войти в аккаунт", 401)
	case errors.Is(err, auth.ErrUnverified):
		jsonError(w, "Подтвердите email перед входом", 403)
	case errors.Is(err, auth.ErrConflict):
		jsonError(w, "Email или логин уже используется", 409)
	case errors.Is(err, auth.ErrAlreadyVerified):
		jsonError(w, "Email уже подтверждён", 409)
	case errors.Is(err, auth.ErrInvalidToken):
		jsonError(w, "Ссылка подтверждения недействительна или истекла", 400)
	case errors.Is(err, auth.ErrDelivery):
		log.Printf("Email delivery failed: %v", err)
		jsonError(w, "Не удалось отправить письмо. Повторите отправку позже", 503)
	default:
		log.Printf("ERROR: %s %s: %v", r.Method, r.URL.Path, err)
		jsonError(w, "Внутренняя ошибка сервера", 500)
	}
}
func decodeRequest[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		jsonError(w, "Ожидается Content-Type: application/json", 415)
		return nil, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input *T
	if err = decoder.Decode(&input); err != nil {
		decodeError(w, err)
		return nil, false
	}
	if input == nil {
		jsonError(w, "Ожидается JSON-объект", 400)
		return nil, false
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = errors.New("extra JSON value")
		}
		decodeError(w, err)
		return nil, false
	}
	return input, true
}
func decodeError(w http.ResponseWriter, err error) {
	var large *http.MaxBytesError
	if errors.As(err, &large) {
		jsonError(w, "Тело запроса слишком большое", 413)
	} else {
		jsonError(w, "Некорректный JSON-запрос", 400)
	}
}
func userResponse(user *models.User) map[string]any {
	return map[string]any{"user": map[string]any{"id": user.ID, "email": user.Email, "login": user.Login, "character_id": user.CharacterID}}
}
