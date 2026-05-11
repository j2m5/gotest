package handlers

import (
	"gotest/internal/db"
	"gotest/internal/templates"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		files := []string{
			"templates/layout.html",
			"templates/register.html",
		}

		data := map[string]any{
			"Title": "Register",
		}

		templates.Render(w, files, data)

		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := db.CreateUser(db.Pool, username, password, "user")

		if err != nil {
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}
}
