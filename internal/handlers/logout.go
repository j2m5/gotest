package handlers

import (
	"net/http"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	cookie, err := r.Cookie("session_token")

	if err != nil {
		h.storage.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	jsonResponse(w, map[string]string{
		"message": "Logout",
	}, http.StatusOK)
}
