package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"gotest/internal/auth"
	"gotest/internal/validation"
)

type resetRecord struct {
	userID           int
	created, expires time.Time
}

func (m *memoryRepository) CreatePasswordReset(ctx context.Context, email, token string, expires time.Time) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if m.err != nil {
		return false, m.err
	}
	u, ok := m.users[email]
	if !ok {
		return false, nil
	}
	if m.resets == nil {
		m.resets = make(map[string]resetRecord)
	}
	for key, value := range m.resets {
		if value.userID != u.User.ID {
			continue
		}
		if value.created.Add(time.Minute).After(time.Now()) {
			return false, nil
		}
		delete(m.resets, key)
	}
	m.resets[token] = resetRecord{u.User.ID, time.Now(), expires}
	return true, nil
}

func (m *memoryRepository) ResetPassword(ctx context.Context, token, password string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if m.err != nil {
		return "", m.err
	}
	record, ok := m.resets[token]
	if !ok || !record.expires.After(time.Now()) {
		return "", auth.ErrInvalidResetToken
	}
	for email, user := range m.users {
		if user.User.ID != record.userID {
			continue
		}
		user.PasswordHash = password
		user.User.CharacterID = nil
		for key, session := range m.sessions {
			if session.User.ID == record.userID {
				delete(m.sessions, key)
			}
		}
		delete(m.resets, token)
		return email, nil
	}
	return "", auth.ErrInvalidResetToken
}

func (m *recordingSender) SendPasswordReset(ctx context.Context, email, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.resetCount++
	m.resetToken = token
	return m.err
}
func (m *recordingSender) SendPasswordChanged(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.changedCount++
	return m.err
}

func TestPasswordResetLifecycleAndIndependentTokens(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		s, repo, mail := fixture(t)
		registerFixture(t, s)
		if err := s.RequestPasswordReset(ctx, "unknown@example.test"); err != nil || mail.resetCount != 0 {
			t.Fatal("unknown account sent mail")
		}
		if err := s.RequestPasswordReset(ctx, " Person@Example.test "); err != nil {
			t.Fatal(err)
		}
		token := mail.resetToken
		stored, ok := repo.resets[auth.TokenDigest(token)]
		if !ok || !stored.expires.Equal(time.Now().Add(30*time.Minute)) || len(token) != 64 {
			t.Fatal("reset token or default expiry incorrect")
		}
		if _, ok := repo.resets[token]; ok {
			t.Fatal("stored raw token")
		}
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil || mail.resetCount != 1 {
			t.Fatal("cooldown bypassed")
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.token, Password: "new password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("verification token reset password")
		}
		if err := s.VerifyEmail(ctx, token); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatal("reset token verified email")
		}
		for _, password := range []string{"short", strings.Repeat("я", 37)} {
			var invalid *validation.Error
			if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: token, Password: password}); !errors.As(err, &invalid) || len(invalid.Fields["password"]) == 0 {
				t.Fatalf("invalid password accepted: %v", err)
			}
		}
		// Recovery before verification works, but does not verify the address.
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: token, Password: "new password"}); err != nil {
			t.Fatal(err)
		}
		if repo.users["person@example.test"].User.EmailVerifiedAt != nil || mail.changedCount != 1 {
			t.Fatal("reset changed verification or skipped notice")
		}
		if _, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "new password"}); !errors.Is(err, auth.ErrUnverified) {
			t.Fatal("reset bypassed verification")
		}
		if err := s.VerifyEmail(ctx, mail.token); err != nil {
			t.Fatal(err)
		}
		ordinary, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "new password"})
		if err != nil {
			t.Fatal(err)
		}
		remembered, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "new password", RememberMe: true})
		if err != nil {
			t.Fatal(err)
		}
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.resetToken, Password: "final password"}); err != nil {
			t.Fatal(err)
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.resetToken, Password: "replayed password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("reset token replayed")
		}
		for _, token := range []string{ordinary.Token, remembered.Token} {
			if _, err := s.CurrentUser(ctx, token); !errors.Is(err, auth.ErrUnauthenticated) {
				t.Fatal("old session survived reset")
			}
		}
		if _, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "new password"}); !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatal("old password accepted")
		}
		if _, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "final password"}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPasswordResetExpiryReplacementAndDeliveryFailures(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		repo, mail := newMemoryRepository(), &recordingSender{}
		s := auth.NewService(repo, mail, auth.Options{PasswordResetTTL: 2 * time.Minute})
		registerFixture(t, s)
		originalPassword := repo.users["person@example.test"].PasswordHash
		mail.err = errors.New("private SMTP detail")
		if err := s.RequestPasswordReset(ctx, "person@example.test"); !errors.Is(err, auth.ErrDelivery) {
			t.Fatalf("delivery failure lost: %v", err)
		}
		if repo.users["person@example.test"].PasswordHash != originalPassword {
			t.Fatal("request changed password")
		}
		old := mail.resetToken
		time.Sleep(time.Minute)
		mail.err = nil
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: old, Password: "new password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("replacement kept old token")
		}
		time.Sleep(2 * time.Minute)
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.resetToken, Password: "new password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("expiry boundary accepted")
		}
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		mail.err = errors.New("notification unavailable")
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.resetToken, Password: "new password"}); !errors.Is(err, auth.ErrPasswordChangedDelivery) {
			t.Fatalf("notification result: %v", err)
		}
		if repo.users["person@example.test"].PasswordHash == originalPassword {
			t.Fatal("notification failure rolled back reset")
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: mail.resetToken, Password: "another password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("failed notice left reset reusable")
		}
	})
}
