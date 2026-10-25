package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/middleware"
	"math"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeRequest[LoginRequest](w, r)
	if !ok {
		return
	}
	result, err := h.auth.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		h.respondError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.SessionCookie, Value: result.Token, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: result.ExpiresAt.UTC(), MaxAge: int(math.Ceil(time.Until(result.ExpiresAt).Seconds()))})
	jsonResponse(w, userResponse(result.User), 200)
}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		h.respondError(w, r, auth.ErrUnauthenticated)
		return
	}
	jsonResponse(w, userResponse(user), 200)
}
