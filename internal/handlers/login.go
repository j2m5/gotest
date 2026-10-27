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
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	input, ok := httpapi.DecodeJSON[LoginRequest](w, r)
	if !ok {
		return
	}
	result, err := h.auth.Login(r.Context(), auth.LoginInput{Email: input.Email, Password: input.Password, RememberMe: input.RememberMe})
	if err != nil {
		respondError(w, r, err)
		return
	}
	cookie := &http.Cookie{Name: middleware.SessionCookie, Value: result.Token, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode}
	if result.Persistent {
		cookie.Expires = result.ExpiresAt.UTC()
		cookie.MaxAge = int(math.Ceil(time.Until(result.ExpiresAt).Seconds()))
		// A deadline reached while persisting the session must not turn into
		// MaxAge=0, which would accidentally produce a browser-session cookie.
		if cookie.MaxAge <= 0 {
			cookie.MaxAge = -1
		}
	}
	http.SetCookie(w, cookie)
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
