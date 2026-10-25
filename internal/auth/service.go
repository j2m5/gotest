package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gotest/internal/models"
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

func NewService(repository Repository, sender VerificationSender, options Options) *Service {
	if options.SessionTTL <= 0 {
		options.SessionTTL = 7 * 24 * time.Hour
	}
	if options.VerificationTTL <= 0 {
		options.VerificationTTL = 24 * time.Hour
	}
	return &Service{repository: repository, sender: sender, options: options, now: time.Now, passwordCost: bcrypt.DefaultCost}
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
	if err = s.repository.CreateRegistration(ctx, input.Email, input.Login, string(password), TokenDigest(token), s.now().Add(s.options.VerificationTTL)); err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return s.deliver(ctx, input.Email, token)
}
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.checkCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}
	if user.EmailVerifiedAt == nil {
		return nil, ErrUnverified
	}
	token := newToken()
	expires := s.now().Add(s.options.SessionTTL)
	if err = s.repository.CreateSession(ctx, user.ID, TokenDigest(token), expires); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &LoginResult{User: user, Token: token, ExpiresAt: expires}, nil
}
func (s *Service) CurrentUser(ctx context.Context, token string) (*models.User, error) {
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
	user, err := s.checkCredentials(ctx, email, password)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}
	token := newToken()
	if err = s.repository.ReplaceVerification(ctx, user.ID, TokenDigest(token), s.now().Add(s.options.VerificationTTL)); err != nil {
		return fmt.Errorf("replace verification: %w", err)
	}
	return s.deliver(ctx, user.Email, token)
}
func (s *Service) checkCredentials(ctx context.Context, email, password string) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(password) == 0 || len(password) > 72 {
		return nil, ErrInvalidCredentials
	}
	user, err := s.repository.FindUserByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("check stored password: %w", err)
	}
	return user, nil
}
func (s *Service) deliver(ctx context.Context, email, token string) error {
	if err := s.sender.SendVerification(ctx, email, token); err != nil {
		return errors.Join(ErrDelivery, fmt.Errorf("send verification: %w", err))
	}
	return nil
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
