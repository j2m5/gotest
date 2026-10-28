package middleware

import (
	"context"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/httpapi"
	"gotest/internal/observability"
	"log/slog"
	"net/http"
)

func (m *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookie)
		if err != nil {
			httpapi.WriteError(w, 401, httpapi.CodeUnauthenticated, "Необходимо войти в аккаунт", nil)
			return
		}
		user, err := m.auth.CurrentUser(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrUnauthenticated) {
			httpapi.WriteError(w, 401, httpapi.CodeUnauthenticated, "Необходимо войти в аккаунт", nil)
			return
		}
		if err != nil || user == nil {
			observability.Logger(r.Context()).LogAttrs(r.Context(), slog.LevelError, "authentication_failed", observability.ErrorAttrs(err)...)
			httpapi.WriteError(w, 500, httpapi.CodeInternalError, "Внутренняя ошибка сервера", nil)
			return
		}
		expected, valid := expectedAccount(w, r)
		if !valid || !accountMatches(w, expected, user.ID) {
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		ctx = context.WithValue(ctx, sessionHashContextKey{}, auth.TokenDigest(cookie.Value))
		next(w, r.WithContext(ctx))
	}
}
