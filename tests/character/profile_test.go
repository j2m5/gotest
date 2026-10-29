package character_test

import (
	"context"
	"errors"
	"testing"

	"gotest/internal/character"
)

func TestPublicProfileRejectsInvalidActorAndIDBeforePersistence(t *testing.T) {
	repo := &repository{}
	service := character.NewService(repo)
	for _, actor := range []character.Actor{{}, {UserID: 7}, {SessionHash: "session"}} {
		if _, err := service.Profile(context.Background(), actor, 1); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("missing actor: %v", err)
		}
	}
	for _, id := range []int{0, -1, 2147483648} {
		if _, err := service.Profile(context.Background(), owner(), id); !errors.Is(err, character.ErrNotFound) {
			t.Fatalf("invalid ID: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Profile(ctx, owner(), 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled profile: %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("invalid input reached storage: %d calls", repo.calls)
	}
	profile, err := service.Profile(context.Background(), owner(), 12)
	if err != nil || profile.ID != 12 || repo.actor != owner() {
		t.Fatalf("target and viewer were mixed: %+v, %+v, %v", profile, repo.actor, err)
	}
	for _, cause := range []error{character.ErrNotFound, character.ErrUnauthenticated, errors.New("storage unavailable")} {
		repo.err = cause
		if _, err := service.Profile(context.Background(), owner(), 1); !errors.Is(err, cause) {
			t.Fatalf("profile error changed: %v", err)
		}
	}
}
