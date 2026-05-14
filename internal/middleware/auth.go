package middleware

import (
	"gotest/internal/auth"
	"net/http"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.CurrentUser(r)

		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)

			return
		}

		next(w, r)
	}
}
