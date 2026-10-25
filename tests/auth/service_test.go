package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gotest/internal/auth"
	"gotest/internal/models"
	"gotest/internal/validation"
)

// This in-memory boundary substitutes only PostgreSQL; the service, password
// hashing, token generation and all authorization decisions remain real.
type memoryRepository struct {
	users         map[string]*models.User
	sessions      map[string]*auth.Session
	verifications map[string]struct {
		userID  int
		expires time.Time
	}
	err             error
	credentialReads int
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{users: map[string]*models.User{}, sessions: map[string]*auth.Session{}, verifications: map[string]struct {
		userID  int
		expires time.Time
	}{}}
}

func (m *memoryRepository) CreateRegistration(ctx context.Context, registration auth.Registration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	for _, u := range m.users {
		if u.Email == registration.Email || u.Login == registration.Login {
			return auth.ErrConflict
		}
	}
	u := &models.User{ID: len(m.users) + 1, Email: registration.Email, Login: registration.Login, Password: registration.PasswordHash}
	m.users[registration.Email] = u
	m.verifications[registration.VerificationTokenHash] = struct {
		userID  int
		expires time.Time
	}{u.ID, registration.VerificationExpiresAt}
	return nil
}
func (m *memoryRepository) FindCredentialsByEmail(ctx context.Context, email string) (*auth.Credentials, error) {
	m.credentialReads++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.users[email]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return &auth.Credentials{User: identityFromRecord(u), PasswordHash: u.Password}, nil
}

func identityFromRecord(u *models.User) auth.Identity {
	return auth.Identity{ID: u.ID, RoleID: u.RoleID, CharacterID: u.CharacterID, Email: u.Email, Login: u.Login, EmailVerifiedAt: u.EmailVerifiedAt}
}
func (m *memoryRepository) CreateSession(ctx context.Context, id int, token string, expires time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	for _, u := range m.users {
		if u.ID == id {
			identity := identityFromRecord(u)
			m.sessions[token] = &auth.Session{User: &identity, ExpiresAt: expires}
			return nil
		}
	}
	return auth.ErrNotFound
}
func (m *memoryRepository) FindSession(ctx context.Context, token string) (*auth.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.sessions[token]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return s, nil
}
func (m *memoryRepository) DeleteSession(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	delete(m.sessions, token)
	return nil
}
func (m *memoryRepository) ConsumeVerification(ctx context.Context, token string, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	v, ok := m.verifications[token]
	if !ok || !v.expires.After(now) {
		return auth.ErrInvalidToken
	}
	for _, u := range m.users {
		if u.ID == v.userID {
			u.EmailVerifiedAt = &now
		}
	}
	for k, other := range m.verifications {
		if other.userID == v.userID {
			delete(m.verifications, k)
		}
	}
	return nil
}
func (m *memoryRepository) ReplaceVerification(ctx context.Context, id int, token string, expires time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	for _, u := range m.users {
		if u.ID == id && u.EmailVerifiedAt != nil {
			return auth.ErrAlreadyVerified
		}
	}
	for k, v := range m.verifications {
		if v.userID == id {
			delete(m.verifications, k)
		}
	}
	m.verifications[token] = struct {
		userID  int
		expires time.Time
	}{id, expires}
	return nil
}

type recordingSender struct {
	email, token string
	err          error
	count        int
}

func (m *recordingSender) SendVerification(ctx context.Context, email, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.email, m.token = email, token
	m.count++
	return m.err
}
func fixture(t *testing.T) (*auth.Service, *memoryRepository, *recordingSender) {
	t.Helper()
	repo, mail := newMemoryRepository(), &recordingSender{}
	s := auth.NewService(repo, mail, auth.Options{SessionTTL: time.Hour, VerificationTTL: 30 * time.Minute})
	return s, repo, mail
}
func registerFixture(t *testing.T, s *auth.Service) {
	t.Helper()
	if err := s.Register(context.Background(), auth.RegisterInput{Email: " Person@Example.test ", Login: " player ", Password: "good password"}); err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationVerificationSessionLifecycle(t *testing.T) {
	s, repo, mail := fixture(t)
	ctx := context.Background()
	registerFixture(t, s)
	u := repo.users["person@example.test"]
	if u == nil || u.Login != "player" || mail.email != "person@example.test" {
		t.Fatalf("normalization failed: %#v, %#v", u, mail)
	}
	if u.Password == "good password" || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("good password")) != nil {
		t.Fatal("password not securely hashed")
	}
	if _, ok := repo.verifications[mail.token]; ok {
		t.Fatal("raw verification token stored")
	}
	if _, err := s.Login(ctx, u.Email, "good password"); !errors.Is(err, auth.ErrUnverified) {
		t.Fatalf("unverified login: %v", err)
	}
	if err := s.VerifyEmail(ctx, mail.token); err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyEmail(ctx, mail.token); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("reused token: %v", err)
	}
	login, err := s.Login(ctx, " Person@Example.test ", "good password")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.sessions[login.Token]; ok {
		t.Fatal("raw session token stored")
	}
	user, err := s.CurrentUser(ctx, login.Token)
	if err != nil || user.ID != u.ID {
		t.Fatalf("session: %v %v", user, err)
	}
	if err = s.Logout(ctx, login.Token); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CurrentUser(ctx, login.Token); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("revoked session: %v", err)
	}
	if err = s.Logout(ctx, login.Token); err != nil {
		t.Fatalf("repeated logout: %v", err)
	}
	if err = s.Logout(ctx, ""); err != nil {
		t.Fatalf("anonymous logout: %v", err)
	}
}

func TestExpiredTokensAreRejectedAtBoundary(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, _, mail := fixture(t)
		registerFixture(t, s)
		ctx := context.Background()
		// Advance the standard library's test clock without changing service internals.
		time.Sleep(30 * time.Minute)
		if err := s.VerifyEmail(ctx, mail.token); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expired verification: %v", err)
		}
		if err := s.ResendVerification(ctx, "person@example.test", "good password"); err != nil {
			t.Fatal(err)
		}
		if err := s.VerifyEmail(ctx, mail.token); err != nil {
			t.Fatal(err)
		}
		login, err := s.Login(ctx, "person@example.test", "good password")
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Until(login.ExpiresAt))
		if _, err = s.CurrentUser(ctx, login.Token); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatalf("expired session: %v", err)
		}
	})
}

func TestMailFailureCanBeRecoveredWithoutAnotherAccount(t *testing.T) {
	s, repo, mail := fixture(t)
	ctx := context.Background()
	mail.err = errors.New("smtp unavailable")
	err := s.Register(ctx, auth.RegisterInput{Email: "person@example.test", Login: "player", Password: "good password"})
	if !errors.Is(err, auth.ErrDelivery) || len(repo.users) != 1 {
		t.Fatalf("delivery failure must preserve created account: %v", err)
	}
	old := mail.token
	mail.err = nil
	if err = s.ResendVerification(ctx, "person@example.test", "wrong password"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("unauthorized resend: %v", err)
	}
	if err = s.ResendVerification(ctx, "person@example.test", "good password"); err != nil {
		t.Fatal(err)
	}
	if err = s.VerifyEmail(ctx, old); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("old token survived resend: %v", err)
	}
	if err = s.VerifyEmail(ctx, mail.token); err != nil {
		t.Fatal(err)
	}
	if len(repo.users) != 1 {
		t.Fatal("resend duplicated account")
	}
	if err = s.ResendVerification(ctx, "person@example.test", "good password"); !errors.Is(err, auth.ErrAlreadyVerified) {
		t.Fatalf("verified resend: %v", err)
	}
}

func TestRegistrationValidationDoesNotPersistInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input auth.RegisterInput
	}{
		{"email", auth.RegisterInput{Email: "not email", Login: "player", Password: "password"}},
		{"display name", auth.RegisterInput{Email: "Person <p@example.test>", Login: "player", Password: "password"}},
		{"long email", auth.RegisterInput{Email: strings.Repeat("a", 250) + "@b.test", Login: "player", Password: "password"}},
		{"blank login", auth.RegisterInput{Email: "p@example.test", Login: "  ", Password: "password"}},
		{"long login", auth.RegisterInput{Email: "p@example.test", Login: strings.Repeat("я", 33), Password: "password"}},
		{"short password", auth.RegisterInput{Email: "p@example.test", Login: "player", Password: "1234567"}},
		{"long password", auth.RegisterInput{Email: "p@example.test", Login: "player", Password: strings.Repeat("я", 37)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, repo, mail := fixture(t)
			err := s.Register(context.Background(), tc.input)
			var invalid *validation.Error
			if !errors.As(err, &invalid) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if len(repo.users) != 0 || mail.count != 0 {
				t.Fatal("invalid input had side effects")
			}
		})
	}
}

func TestValidLongEmailAndUnicodeLogin(t *testing.T) {
	s, repo, _ := fixture(t)
	email := "long.valid.email.address@example.test"
	if err := s.Register(context.Background(), auth.RegisterInput{Email: email, Login: strings.Repeat("я", 32), Password: strings.Repeat("я", 36)}); err != nil {
		t.Fatal(err)
	}
	if repo.users[email] == nil {
		t.Fatal("valid long address rejected")
	}
}

func TestErrorsAreNotDisguisedAsInvalidCredentials(t *testing.T) {
	s, repo, mail := fixture(t)
	ctx := context.Background()
	registerFixture(t, s)
	if err := s.Register(ctx, auth.RegisterInput{Email: "person@example.test", Login: "another", Password: "password"}); !errors.Is(err, auth.ErrConflict) {
		t.Fatalf("duplicate: %v", err)
	}
	if _, err := s.Login(ctx, "missing@example.test", "password"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("missing user: %v", err)
	}
	if _, err := s.Login(ctx, "person@example.test", "wrong password"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("bad password: %v", err)
	}
	if err := s.VerifyEmail(ctx, mail.token); err != nil {
		t.Fatal(err)
	}
	login, err := s.Login(ctx, "person@example.test", "good password")
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("database unavailable")
	repo.err = failure
	if _, err = s.Login(ctx, "person@example.test", "good password"); !errors.Is(err, failure) {
		t.Fatalf("db login error lost: %v", err)
	}
	if _, err = s.CurrentUser(ctx, login.Token); !errors.Is(err, failure) {
		t.Fatalf("db session error lost: %v", err)
	}
	if err = s.Logout(ctx, login.Token); !errors.Is(err, failure) {
		t.Fatalf("failed revoke reported success: %v", err)
	}
}

func TestCancelledRegistrationHasNoSideEffects(t *testing.T) {
	s, repo, mail := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.Register(ctx, auth.RegisterInput{Email: "p@example.test", Login: "player", Password: "password"})
	if !errors.Is(err, context.Canceled) || len(repo.users) != 0 || mail.count != 0 {
		t.Fatalf("cancelled registration: %v", err)
	}
}
