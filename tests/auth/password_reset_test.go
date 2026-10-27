package auth_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"gotest/internal/auth"
	"gotest/internal/notification"
	"gotest/internal/validation"
)

type resetRecord struct {
	userID           int
	created, expires time.Time
}

func (m *memoryRepository) EnqueuePasswordReset(ctx context.Context, email string, expires time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	m.requests = append(m.requests, auth.QueuedResetRequest{ID: fmt.Sprintf("request-%d", len(m.requestExpires)+1), Lease: "current-lease", Email: email})
	m.requestExpires = append(m.requestExpires, expires)
	return nil
}

func (m *memoryRepository) CreatePasswordReset(ctx context.Context, input auth.PasswordReset) (bool, error) {
	email, token, expires := input.Email, input.TokenHash, input.ExpiresAt
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if m.err != nil {
		return false, m.err
	}
	index := -1
	for i, request := range m.requests {
		if request.ID == input.RequestID && request.Lease == input.RequestLease && request.Email == input.Email {
			index = i
			break
		}
	}
	if index < 0 {
		return false, notification.ErrLeaseLost
	}
	consume := func() { m.requests = append(m.requests[:index], m.requests[index+1:]...) }
	u, ok := m.users[email]
	if !ok {
		consume()
		return false, nil
	}
	if m.resets == nil {
		m.resets = make(map[string]resetRecord)
	}
	for _, value := range m.resets {
		if value.userID != u.User.ID {
			continue
		}
		if value.created.Add(time.Minute).After(time.Now()) {
			consume()
			return false, nil
		}
	}
	if err := m.mail.SendPasswordReset(ctx, email, input.Token); err != nil {
		return false, err
	}
	for key, value := range m.resets {
		if value.userID == u.User.ID {
			delete(m.resets, key)
		}
	}
	m.resets[token] = resetRecord{u.User.ID, time.Now(), expires}
	consume()
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
		if err := m.mail.SendPasswordChanged(ctx, email); err != nil {
			return "", err
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
	if m.err != nil {
		return m.err
	}
	m.resetCount++
	m.resetToken = token
	return m.err
}
func (m *recordingSender) SendPasswordChanged(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	m.changedCount++
	return m.err
}

func drainResetRequests(t *testing.T, s *auth.Service, repo *memoryRepository) {
	t.Helper()
	for len(repo.requests) > 0 {
		if err := s.ProcessPasswordReset(context.Background(), repo.requests[0]); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPasswordResetLifecycleAndIndependentTokens(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		s, repo, mail := fixture(t)
		registerFixture(t, s)
		if err := s.RequestPasswordReset(ctx, "unknown@example.test"); err != nil || mail.resetCount != 0 {
			t.Fatal("unknown account sent mail")
		}
		drainResetRequests(t, s, repo)
		if err := s.RequestPasswordReset(ctx, " Person@Example.test "); err != nil {
			t.Fatal(err)
		}
		drainResetRequests(t, s, repo)
		token := mail.resetToken
		stored, ok := repo.resets[auth.TokenDigest(token)]
		if !ok || !stored.expires.Equal(time.Now().Add(30*time.Minute)) || len(token) != 64 {
			t.Fatal("reset token or default expiry incorrect")
		}
		if _, ok := repo.resets[token]; ok {
			t.Fatal("stored raw token")
		}
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		drainResetRequests(t, s, repo)
		if mail.resetCount != 1 {
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
		drainResetRequests(t, s, repo)
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

func TestPasswordResetExpiryReplacementAndOutboxRollback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		repo, mail := newMemoryRepository(), &recordingSender{}
		repo.mail = mail
		s := auth.NewService(repo, auth.Options{PasswordResetTTL: 2 * time.Minute})
		registerFixture(t, s)
		originalPassword := repo.users["person@example.test"].PasswordHash
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		drainResetRequests(t, s, repo)
		old := mail.resetToken
		time.Sleep(time.Minute)
		if err := s.RequestPasswordReset(ctx, "person@example.test"); err != nil {
			t.Fatal(err)
		}
		request := repo.requests[0]
		failure := errors.New("outbox persistence unavailable")
		mail.err = failure
		if err := s.ProcessPasswordReset(ctx, request); !errors.Is(err, failure) {
			t.Fatalf("outbox failure lost: %v", err)
		}
		if _, ok := repo.resets[auth.TokenDigest(old)]; !ok || len(repo.requests) != 1 || repo.users["person@example.test"].PasswordHash != originalPassword {
			t.Fatal("failed outgoing enqueue consumed source or changed account")
		}
		mail.err = nil
		drainResetRequests(t, s, repo)
		if err := s.ProcessPasswordReset(ctx, request); !errors.Is(err, notification.ErrLeaseLost) {
			t.Fatal("consumed request replayed")
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
		drainResetRequests(t, s, repo)
		token := mail.resetToken
		// Include a live session so failed notice persistence cannot silently revoke it.
		if err := s.VerifyEmail(ctx, mail.token); err != nil {
			t.Fatal(err)
		}
		login, err := s.Login(ctx, auth.LoginInput{Email: "person@example.test", Password: "good password"})
		if err != nil {
			t.Fatal(err)
		}
		mail.err = failure
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: token, Password: "new password"}); !errors.Is(err, failure) {
			t.Fatalf("notice enqueue failure lost: %v", err)
		}
		if repo.users["person@example.test"].PasswordHash != originalPassword || len(repo.resets) != 1 || mail.changedCount != 0 {
			t.Fatal("failed notice changed password or consumed token")
		}
		if _, err := s.CurrentUser(ctx, login.Token); err != nil {
			t.Fatal("failed notice revoked session")
		}
		mail.err = nil
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: token, Password: "new password"}); err != nil {
			t.Fatal(err)
		}
		if err := s.ResetPassword(ctx, auth.ResetPasswordInput{Token: token, Password: "another password"}); !errors.Is(err, auth.ErrInvalidResetToken) {
			t.Fatal("successful reset left token reusable")
		}
		if _, err := s.CurrentUser(ctx, login.Token); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatal("successful reset kept session")
		}
	})
}

func TestResetIntakeOnlyEnqueuesNormalizedKnownAndUnknownAddresses(t *testing.T) {
	ctx := context.Background()
	s, repo, mail := fixture(t)
	registerFixture(t, s)
	for _, email := range []string{" Person@Example.test ", " Unknown@Example.test "} {
		if err := s.RequestPasswordReset(ctx, email); err != nil {
			t.Fatal(err)
		}
	}
	if repo.credentialReads != 0 || len(repo.resets) != 0 || mail.resetCount != 0 || len(repo.requests) != 2 {
		t.Fatal("intake looked up accounts or generated outgoing token")
	}
	if repo.requests[0].Email != "person@example.test" || repo.requests[1].Email != "unknown@example.test" {
		t.Fatal("intake failed normalization")
	}
	if !repo.requestExpires[0].After(time.Now()) {
		t.Fatal("queued request lacks expiry")
	}
	drainResetRequests(t, s, repo)
	if mail.resetCount != 1 || len(repo.resets) != 1 || len(repo.requests) != 0 {
		t.Fatal("processing failed known/unknown account handling")
	}
	failure := errors.New("inbox unavailable")
	repo.err = failure
	if err := s.RequestPasswordReset(ctx, "unknown@example.test"); !errors.Is(err, failure) {
		t.Fatalf("intake persistence failure lost: %v", err)
	}
	if len(repo.requests) != 0 || repo.credentialReads != 0 {
		t.Fatal("failed enqueue had side effects")
	}
}

func TestResetIntakeRejectsInvalidAndCancelledRequests(t *testing.T) {
	s, repo, _ := fixture(t)
	for _, email := range []string{"", "bad", "Name <user@example.test>", "user@example.test\r\nBcc: other@example.test"} {
		var invalid *validation.Error
		if err := s.RequestPasswordReset(context.Background(), email); !errors.As(err, &invalid) {
			t.Fatalf("invalid intake accepted: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.RequestPasswordReset(ctx, "user@example.test"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error lost: %v", err)
	}
	if len(repo.requests) != 0 || repo.credentialReads != 0 {
		t.Fatal("rejected request had side effects")
	}
}
