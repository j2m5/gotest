package middleware

import (
	"net/http"
)

func (m *Middleware) Guest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := m.auth.CurrentUser(r)

		if user != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}

		next(w, r)
	}
}
