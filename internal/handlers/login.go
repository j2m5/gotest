package handlers

import (
	"gotest/internal/auth"
	"gotest/internal/httpapi"
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
	input, ok := httpapi.DecodeJSON[LoginRequest](w, r)
	if !ok {
		return
	}
	result, err := h.auth.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		respondError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.SessionCookie, Value: result.Token, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: result.ExpiresAt.UTC(), MaxAge: int(math.Ceil(time.Until(result.ExpiresAt).Seconds()))})
	httpapi.WriteJSON(w, 200, userResponse(result.User))
}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respondError(w, r, auth.ErrUnauthenticated)
		return
	}
	httpapi.WriteJSON(w, 200, userResponse(user))
}
