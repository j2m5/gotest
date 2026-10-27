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
	if !validEmail(input.Email) {
		fields["email"] = []string{"Укажите корректный email длиной до 254 байт"}
	}
	length := utf8.RuneCountInString(input.Login)
	if length < 1 || length > 32 || strings.IndexFunc(input.Login, unicode.IsControl) >= 0 {
		fields["login"] = []string{"Логин должен содержать от 1 до 32 символов без управляющих символов"}
	}
	if message := passwordError(input.Password); message != "" {
		fields["password"] = []string{message}
	}
	if len(fields) > 0 {
		return &validation.Error{Fields: fields}
	}
	return nil
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email && len(email) <= 254 && !strings.ContainsAny(email, "\r\n")
}

func passwordError(password string) string {
	if len(password) < 8 || len(password) > 72 {
		return "Пароль должен содержать от 8 до 72 байт"
	}
	return ""
}

// ValidatePasswordResetEmail performs no account lookup, so accepting a request
// can be independent of whether that account exists and how long mail takes.
func ValidatePasswordResetEmail(email string) (string, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return "", &validation.Error{Fields: validation.Fields{"email": {"Укажите корректный email длиной до 254 байт"}}}
	}
	return email, nil
}
func validToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}
