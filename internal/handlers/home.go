package handlers

import (
	"gotest/internal/templates"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"templates/layout.html",
		"templates/home.html",
	}

	data := map[string]any{
		"Title": "Go Test",
	}

	templates.Render(w, files, data)
}
