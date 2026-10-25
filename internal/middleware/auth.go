package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"gotest/internal/auth"
	"log"
	"net/http"
)

func (m *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookie)
		if err != nil {
			writeError(w, 401, "Необходимо войти в аккаунт")
			return
		}
		user, err := m.auth.CurrentUser(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrUnauthenticated) {
			writeError(w, 401, "Необходимо войти в аккаунт")
			return
		}
		if err != nil || user == nil {
			log.Printf("Authentication failed: %v", err)
			writeError(w, 500, "Внутренняя ошибка сервера")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		ctx = context.WithValue(ctx, sessionHashContextKey{}, auth.TokenDigest(cookie.Value))
		next(w, r.WithContext(ctx))
	}
}
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
