package auth

import (
	"encoding/hex"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"gotest/internal/validation"
)

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
func validateRegistration(input RegisterInput) error {
	fields := validation.Fields{}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 || strings.ContainsAny(input.Email, "\r\n") {
		fields["email"] = []string{"Укажите корректный email длиной до 254 байт"}
	}
	length := utf8.RuneCountInString(input.Login)
	if length < 1 || length > 32 || strings.IndexFunc(input.Login, unicode.IsControl) >= 0 {
		fields["login"] = []string{"Логин должен содержать от 1 до 32 символов без управляющих символов"}
	}
	if len(input.Password) < 8 || len(input.Password) > 72 {
		fields["password"] = []string{"Пароль должен содержать от 8 до 72 байт"}
	}
	if len(fields) > 0 {
		return &validation.Error{Fields: fields}
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
