package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// A valid production-cost hash gives missing users the same expensive password check.
var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("unused authentication timing password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

func NewService(repository Repository, options Options) *Service {
	if options.SessionTTL <= 0 {
		options.SessionTTL = 24 * time.Hour
	}
	if options.RememberMeTTL <= 0 {
		options.RememberMeTTL = 30 * 24 * time.Hour
	}
	if options.VerificationTTL <= 0 {
		options.VerificationTTL = 24 * time.Hour
	}
	if options.PasswordResetTTL <= 0 {
		options.PasswordResetTTL = 30 * time.Minute
	}
	return &Service{repository: repository, options: options, now: time.Now, passwordCost: bcrypt.DefaultCost}
}
func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	input.Email = normalizeEmail(input.Email)
	input.Login = strings.TrimSpace(input.Login)
	if err := validateRegistration(input); err != nil {
		return err
	}
	password, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.passwordCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	token := newToken()
	registration := Registration{
		Email:                 input.Email,
		Login:                 input.Login,
		PasswordHash:          string(password),
		VerificationTokenHash: TokenDigest(token),
		VerificationToken:     token,
		VerificationExpiresAt: s.now().Add(s.options.VerificationTTL),
	}
	if err = s.repository.CreateRegistration(ctx, registration); err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return nil
}
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	credentials, err := s.checkCredentials(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	user := &credentials.User
	if user.EmailVerifiedAt == nil {
		return nil, ErrUnverified
	}
	token := newToken()
	ttl := s.options.SessionTTL
	if input.RememberMe {
		ttl = s.options.RememberMeTTL
	}
	expires := s.now().Add(ttl)
	if err = s.repository.CreateSession(ctx, SessionInput{UserID: user.ID, PasswordHash: credentials.PasswordHash, TokenHash: TokenDigest(token), ExpiresAt: expires}); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &LoginResult{User: user, Token: token, ExpiresAt: expires, Persistent: input.RememberMe}, nil
}
func (s *Service) CurrentUser(ctx context.Context, token string) (*Identity, error) {
	if !validToken(token) {
		return nil, ErrUnauthenticated
	}
	session, err := s.repository.FindSession(ctx, TokenDigest(token))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthenticated
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if !session.ExpiresAt.After(s.now()) || session.User.EmailVerifiedAt == nil {
		return nil, ErrUnauthenticated
	}
	return session.User, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	if !validToken(token) {
		return nil
	}
	if err := s.repository.DeleteSession(ctx, TokenDigest(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	if !validToken(token) {
		return ErrInvalidToken
	}
	return s.repository.ConsumeVerification(ctx, TokenDigest(token), s.now())
}
func (s *Service) ResendVerification(ctx context.Context, email, password string) error {
	credentials, err := s.checkCredentials(ctx, email, password)
	if err != nil {
		return err
	}
	user := &credentials.User
	if user.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}
	token := newToken()
	if err = s.repository.ReplaceVerification(ctx, Verification{UserID: user.ID, Token: token, TokenHash: TokenDigest(token), ExpiresAt: s.now().Add(s.options.VerificationTTL)}); err != nil {
		return fmt.Errorf("replace verification: %w", err)
	}
	return nil
}
func (s *Service) checkCredentials(ctx context.Context, email, password string) (*Credentials, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(password) == 0 || len(password) > 72 {
		return nil, ErrInvalidCredentials
	}
	credentials, err := s.repository.FindCredentialsByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("check stored password: %w", err)
	}
	return credentials, nil
}
func newToken() string {
	var token [32]byte
	// Go 1.26 crypto/rand.Read fills the buffer or terminates the process.
	_, _ = rand.Read(token[:])
	return hex.EncodeToString(token[:])
}

// TokenDigest derives the one-way storage key used for session and verification
// tokens. Hashing alone does not validate a token or authenticate its owner.
func TokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
