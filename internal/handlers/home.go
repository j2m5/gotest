package handlers

import (
	"gotest/internal/templates"
	"net/http"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"templates/layout.html",
		"templates/home.html",
	}

	user := h.auth.CurrentUser(r)

	data := map[string]any{
		"Title":   "Go Test",
		"User":    user,
		"Success": h.flash.Get(w, r, "success"),
		"Errors":  h.flash.Get(w, r, "error"),
	}

	templates.Render(w, files, data)
}
