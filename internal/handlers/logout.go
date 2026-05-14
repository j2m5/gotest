package handlers

import (
	"gotest/internal/db"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")

	if err == nil {
		db.DeleteSession(db.Pool, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
