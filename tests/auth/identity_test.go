package auth_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoginAndSessionResultsDoNotContainCredentials(t *testing.T) {
	s, repo, mail := fixture(t)
	registerFixture(t, s)
	ctx := context.Background()
	if err := s.VerifyEmail(ctx, mail.token); err != nil {
		t.Fatal(err)
	}
	login, err := s.Login(ctx, "person@example.test", "good password")
	if err != nil {
		t.Fatal(err)
	}
	credentialReads := repo.credentialReads
	current, err := s.CurrentUser(ctx, login.Token)
	if err != nil {
		t.Fatal(err)
	}
	if repo.credentialReads != credentialReads {
		t.Fatal("checking a session loaded password credentials")
	}
	if current.ID != login.User.ID || current.Email != login.User.Email || current.Login != "player" || current.CharacterID != nil || current.EmailVerifiedAt == nil {
		t.Fatal("safe identity lost account fields")
	}
	for label, user := range map[string]any{"login": login.User, "current user": current} {
		data, err := json.Marshal(user)
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"password", "hash", "token", "$2a$", "$2b$"} {
			if strings.Contains(strings.ToLower(string(data)), secret) {
				t.Errorf("%s exposes credential field/data: %s", label, secret)
			}
		}
	}
}
