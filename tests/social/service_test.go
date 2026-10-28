package social_test

import (
	"context"
	"errors"
	"gotest/internal/social"
	"gotest/internal/validation"
	"strings"
	"testing"
)

func TestSocialRejectsInvalidTextAndDuration(t *testing.T) {
	s := social.NewService(nil)
	a := social.Actor{UserID: 1, SessionHash: "session"}
	for _, body := range []string{" ", strings.Repeat("я", 1001), string([]byte{255})} {
		_, err := s.Send(context.Background(), a, social.SendInput{Body: body, CharacterID: 1})
		var validationError *validation.Error
		if !errors.As(err, &validationError) {
			t.Fatalf("body validation=%v", err)
		}
	}
	for _, hours := range []int{0, 169} {
		_, err := s.Ban(context.Background(), a, social.BanInput{UserID: 2, Scope: "chat", Reason: "reason", DurationHours: &hours})
		var validationError *validation.Error
		if !errors.As(err, &validationError) {
			t.Fatalf("duration validation=%v", err)
		}
	}
}
