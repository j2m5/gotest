package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gotest/internal/db"
	"gotest/internal/templates"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
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

		user, err := db.FindUserByEmail(db.Pool, email)

		if err != nil {
			fmt.Println(email)
			fmt.Println(password)
			fmt.Println(err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)

			return
		}

		if user.Password != password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)

			return
		}

		token := generateToken()

		err = db.CreateSession(db.Pool, user.ID, token)

		if err != nil {
			http.Error(w, "Cannot create session", http.StatusInternalServerError)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: token,
			Path:  "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateToken() string {
	bytes := make([]byte, 32)

	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}
