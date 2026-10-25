package handlers

import (
	"gotest/internal/middleware"
	"net/http"
	"time"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := ""
	if cookie, err := r.Cookie(middleware.SessionCookie); err == nil {
		token = cookie.Value
	}
	if err := h.auth.Logout(r.Context(), token); err != nil {
		h.respondError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.SessionCookie, Value: "", Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0).UTC()})
	jsonResponse(w, map[string]string{"message": "Вы вышли из аккаунта"}, 200)
}
