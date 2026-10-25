package auth

import (
	"encoding/hex"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
func validateRegistration(input RegisterInput) error {
	var messages []string
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 || strings.ContainsAny(input.Email, "\r\n") {
		messages = append(messages, "Укажите корректный email длиной до 254 байт")
	}
	length := utf8.RuneCountInString(input.Login)
	if length < 1 || length > 32 || strings.IndexFunc(input.Login, unicode.IsControl) >= 0 {
		messages = append(messages, "Логин должен содержать от 1 до 32 символов без управляющих символов")
	}
	if len(input.Password) < 8 || len(input.Password) > 72 {
		messages = append(messages, "Пароль должен содержать от 8 до 72 байт")
	}
	if len(messages) > 0 {
		return &ValidationError{Messages: messages}
	}
	return nil
}
func validToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}
