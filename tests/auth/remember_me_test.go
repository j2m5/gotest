package auth_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"gotest/internal/auth"
)

func TestLoginSessionLifetimesAndRevocation(t *testing.T) {
	for _, tc := range []struct {
		name               string
		options            auth.Options
		normal, remembered time.Duration
	}{
		{"defaults", auth.Options{}, 24 * time.Hour, 30 * 24 * time.Hour},
		{"configured", auth.Options{SessionTTL: time.Hour, RememberMeTTL: 8 * time.Hour}, time.Hour, 8 * time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				repo, sender := newMemoryRepository(), &recordingSender{}
				s := auth.NewService(repo, sender, tc.options)
				registerFixture(t, s)
				input := auth.LoginInput{Email: "person@example.test", Password: "good password", RememberMe: true}
				if _, err := s.Login(ctx, input); !errors.Is(err, auth.ErrUnverified) || len(repo.sessions) != 0 {
					t.Fatalf("remembered login bypassed email verification: %v", err)
				}
				if err := s.VerifyEmail(ctx, sender.token); err != nil {
					t.Fatal(err)
				}
				input.Password = "wrong password"
				if _, err := s.Login(ctx, input); !errors.Is(err, auth.ErrInvalidCredentials) || len(repo.sessions) != 0 {
					t.Fatalf("remembered login bypassed password: %v", err)
				}
				input.Password = "good password"
				start := time.Now()
				var tokens [2]string
				for i, ttl := range []time.Duration{tc.normal, tc.remembered} {
					input.RememberMe = i == 1
					result, err := s.Login(ctx, input)
					if err != nil {
						t.Fatal(err)
					}
					if result.Persistent != input.RememberMe || !result.ExpiresAt.Equal(start.Add(ttl)) {
						t.Fatalf("mode %v: persistent=%v, ttl=%v; want %v", input.RememberMe, result.Persistent, result.ExpiresAt.Sub(start), ttl)
					}
					tokens[i] = result.Token
					stored, ok := repo.sessions[auth.TokenDigest(result.Token)]
					if !ok || !stored.ExpiresAt.Equal(result.ExpiresAt) || repo.sessions[result.Token] != nil {
						t.Fatal("incorrect stored session digest or expiry")
					}
					// A separate token can be revoked without affecting the first one.
					revoked, err := s.Login(ctx, input)
					if err != nil {
						t.Fatal(err)
					}
					if err := s.Logout(ctx, revoked.Token); err != nil {
						t.Fatal(err)
					}
					if _, err := s.CurrentUser(ctx, revoked.Token); !errors.Is(err, auth.ErrUnauthenticated) {
						t.Fatal("revoked session accepted")
					}
				}
				time.Sleep(tc.normal - time.Nanosecond)
				for _, token := range tokens {
					if _, err := s.CurrentUser(ctx, token); err != nil {
						t.Fatal(err)
					}
				}
				time.Sleep(time.Nanosecond)
				if _, err := s.CurrentUser(ctx, tokens[0]); !errors.Is(err, auth.ErrUnauthenticated) {
					t.Fatal("ordinary session survived its fixed deadline")
				}
				if _, err := s.CurrentUser(ctx, tokens[1]); err != nil {
					t.Fatal(err)
				}
				time.Sleep(tc.remembered - tc.normal)
				if _, err := s.CurrentUser(ctx, tokens[1]); !errors.Is(err, auth.ErrUnauthenticated) {
					t.Fatal("remembered session survived its fixed deadline")
				}
			})
		})
	}
}
