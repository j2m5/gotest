package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gotest/internal/validation"
)

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	email, err := ValidatePasswordResetEmail(email)
	if err != nil {
		return err
	}
	if err := s.repository.EnqueuePasswordReset(ctx, email, s.now().Add(s.options.PasswordResetTTL)); err != nil {
		return fmt.Errorf("enqueue password reset: %w", err)
	}
	return nil
}

// ProcessPasswordReset runs only from the durable worker. Persistence consumes
// the input lease atomically even for unknown accounts or cooldown suppression.
func (s *Service) ProcessPasswordReset(ctx context.Context, request QueuedResetRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	email, err := ValidatePasswordResetEmail(request.Email)
	if err != nil {
		return err
	}
	token := newToken()
	_, err = s.repository.CreatePasswordReset(ctx, PasswordReset{Email: email, Token: token,
		TokenHash: TokenDigest(token), ExpiresAt: s.now().Add(s.options.PasswordResetTTL),
		RequestID: request.ID, RequestLease: request.Lease})
	if err != nil {
		return fmt.Errorf("process password reset: %w", err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validToken(input.Token) {
		return ErrInvalidResetToken
	}
	if message := passwordError(input.Password); message != "" {
		return &validation.Error{Fields: validation.Fields{"password": {message}}}
	}
	password, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.passwordCost)
	if err != nil {
		return fmt.Errorf("hash reset password: %w", err)
	}
	_, err = s.repository.ResetPassword(ctx, TokenDigest(input.Token), string(password))
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	return nil
}
