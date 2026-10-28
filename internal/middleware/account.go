package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"gotest/internal/auth"
	"gotest/internal/httpapi"
	"gotest/internal/observability"
)

const AccountHeader = "X-Account-ID"

// expectedAccount is an optional precondition, never an authentication source.
// Zero means no precondition was supplied by this client.
func expectedAccount(w http.ResponseWriter, r *http.Request) (int, bool) {
	values, present := r.Header[http.CanonicalHeaderKey(AccountHeader)]
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
		id, err := strconv.ParseInt(values[0], 10, 32)
		if digits && err == nil && id > 0 {
			return int(id), true
		}
	}
	httpapi.WriteError(w, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Некорректный контекст аккаунта. Обновите страницу", nil)
	return 0, false
}

func accountMatches(w http.ResponseWriter, expected, actual int) bool {
	if expected != 0 && expected != actual {
		httpapi.WriteError(w, http.StatusConflict, httpapi.CodeAccountChanged, "Аккаунт изменился. Обновите данные и повторите действие", nil)
		return false
	}
	return true
}

// AccountLogout preserves idempotent logout of an absent/expired session while
// preventing a stale tab from logging out a different, newly signed-in account.
func (m *Middleware) AccountLogout(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expected, valid := expectedAccount(w, r)
		if !valid {
			return
		}
		cookie, err := r.Cookie(SessionCookie)
		if expected == 0 || err != nil {
			next(w, r)
			return
		}
		user, err := m.auth.CurrentUser(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrUnauthenticated) {
			next(w, r)
			return
		}
		if err != nil || user == nil {
			observability.Logger(r.Context()).LogAttrs(r.Context(), slog.LevelError, "authentication_failed", observability.ErrorAttrs(err)...)
			httpapi.WriteError(w, http.StatusInternalServerError, httpapi.CodeInternalError, "Внутренняя ошибка сервера", nil)
			return
		}
		if accountMatches(w, expected, user.ID) {
			next(w, r)
		}
	}
}
