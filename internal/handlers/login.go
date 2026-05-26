package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"gotest/internal/templates"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		files := []string{
			"templates/layout.html",
			"templates/login.html",
		}

		data := map[string]any{
			"Title": "Login",
		}

		templates.Render(w, files, data)

		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := h.storage.FindUserByEmail(email)

		if err != nil {
			h.flash.Set(w, r, "error", "Неверный email или пароль")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

		if err != nil {
			h.flash.Set(w, r, "error", "Неверный email или пароль")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

			return
		}

		token := generateToken()

		err = h.storage.CreateSession(user.ID, token)

		if err != nil {
			h.serverError(w, r, err)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateToken() string {
	bytes := make([]byte, 32)

	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}
