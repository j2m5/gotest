package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.storage.FindUserByEmail(request.Email)

	if err != nil {
		jsonError(w, "Неверный email или пароль", http.StatusUnauthorized)

		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		jsonError(w, "Неверный email или пароль", http.StatusUnauthorized)

		return
	}

	if user.EmailVerifiedAt == nil {
		jsonError(w, "Подтвердите email перед входом", http.StatusForbidden)

		return
	}

	token := generateToken()

	if err = h.storage.CreateSession(user.ID, token); err != nil {
		h.serverError(w, r, err)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	jsonResponse(w, map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"login": user.Login,
		},
	}, http.StatusOK)
}

func generateToken() string {
	bytes := make([]byte, 32)

	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}
