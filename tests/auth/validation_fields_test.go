package auth_test

import (
	"context"
	"errors"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/validation"
)

func TestRegistrationValidationNamesAffectedFields(t *testing.T) {
	s, repo, mail := fixture(t)
	err := s.Register(context.Background(), auth.RegisterInput{Email: "broken-email", Login: " ", Password: "short"})
	var invalid *validation.Error
	if !errors.As(err, &invalid) {
		t.Fatalf("expected field validation, got %v", err)
	}
	if len(invalid.Fields) != 3 {
		t.Fatalf("unexpected fields: %v", invalid.Fields)
	}
	for _, field := range []string{"email", "login", "password"} {
		if len(invalid.Fields[field]) != 1 || invalid.Fields[field][0] == "" {
			t.Errorf("missing error for %s", field)
		}
	}
	if len(repo.users) != 0 || mail.count != 0 {
		t.Fatal("invalid registration had side effects")
	}
}
