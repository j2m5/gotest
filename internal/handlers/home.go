package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/templates"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"templates/layout.html",
		"templates/home.html",
	}

	user := auth.CurrentUser(r)

	data := map[string]any{
		"Title": "Go Test",
		"User":  user,
	}

	templates.Render(w, files, data)
}
