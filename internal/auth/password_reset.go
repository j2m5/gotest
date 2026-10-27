package auth

import (
	"context"
	"errors"
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
	token := newToken()
	created, err := s.repository.CreatePasswordReset(ctx, email, TokenDigest(token), s.now().Add(s.options.PasswordResetTTL))
	if err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	if !created {
		return nil
	}
	if err := s.sender.SendPasswordReset(ctx, email, token); err != nil {
		return errors.Join(ErrDelivery, fmt.Errorf("send password reset: %w", err))
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
	email, err := s.repository.ResetPassword(ctx, TokenDigest(input.Token), string(password))
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := s.sender.SendPasswordChanged(ctx, email); err != nil {
		return errors.Join(ErrPasswordChangedDelivery, err)
	}
	return nil
}
