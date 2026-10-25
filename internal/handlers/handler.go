package handlers

import (
	"context"
	"gotest/internal/auth"
)

type AuthService interface {
	Register(context.Context, auth.RegisterInput) error
	Login(context.Context, string, string) (*auth.LoginResult, error)
	Logout(context.Context, string) error
	VerifyEmail(context.Context, string) error
	ResendVerification(context.Context, string, string) error
}
type Handler struct {
	auth          AuthService
	secureCookies bool
}

func NewHandler(service AuthService, secureCookies bool) *Handler {
	return &Handler{auth: service, secureCookies: secureCookies}
}

type userDTO struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Login       string `json:"login"`
	CharacterID *int   `json:"character_id"`
}

type accountResponse struct {
	User userDTO `json:"user"`
}

func userResponse(user *auth.Identity) accountResponse {
	return accountResponse{User: userDTO{ID: user.ID, Email: user.Email, Login: user.Login, CharacterID: user.CharacterID}}
}
