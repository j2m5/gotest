package progression_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gotest/internal/progression"
)

type repositoryFunc func(context.Context, progression.Reward) (*progression.Result, error)

func (f repositoryFunc) Grant(ctx context.Context, r progression.Reward) (*progression.Result, error) {
	return f(ctx, r)
}

func TestProgressionServiceRejectsInvalidRewardsBeforeStorage(t *testing.T) {
	calls := 0
	service := progression.NewService(repositoryFunc(func(context.Context, progression.Reward) (*progression.Result, error) {
		calls++
		return nil, nil
	}))
	for _, reward := range []progression.Reward{
		{CharacterID: 0, Key: "quest:1", Experience: 1}, {CharacterID: -1, Key: "quest:1", Experience: 1}, {CharacterID: 1, Key: "quest:1", Experience: 0}, {CharacterID: 1, Key: "quest:1", Experience: -1},
		{CharacterID: 1, Key: "", Experience: 1}, {CharacterID: 1, Key: strings.Repeat("a", 129), Experience: 1}, {CharacterID: 1, Key: " quest:1", Experience: 1}, {CharacterID: 1, Key: "quest:1 ", Experience: 1},
		{CharacterID: 1, Key: "quest\n1", Experience: 1}, {CharacterID: 1, Key: "quest\x001", Experience: 1}, {CharacterID: 1, Key: ":quest", Experience: 1}, {CharacterID: 1, Key: "quest?1", Experience: 1}, {CharacterID: 1, Key: "задание:1", Experience: 1},
	} {
		if _, err := service.Grant(context.Background(), reward); !errors.Is(err, progression.ErrInvalidReward) {
			t.Errorf("reward %+v: error = %v", reward, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Grant(ctx, progression.Reward{CharacterID: 1, Key: "quest:1", Experience: 100}); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled reward error = %v", err)
	}
	if calls != 0 {
		t.Errorf("invalid/cancelled rewards reached storage %d times", calls)
	}
}

func TestProgressionServicePreservesStorageResultAndError(t *testing.T) {
	wantResult := &progression.Result{CharacterID: 1, Replayed: true}
	wantErr := errors.New("storage failed")
	ctx := context.Background()
	for _, key := range []string{"a", "Quest:123/attempt_7.1-2", strings.Repeat("9", 128)} {
		reward := progression.Reward{CharacterID: 1, Key: key, Experience: 100}
		service := progression.NewService(repositoryFunc(func(gotCtx context.Context, got progression.Reward) (*progression.Result, error) {
			if gotCtx != ctx || got != reward {
				t.Errorf("request changed: %+v", got)
			}
			return wantResult, wantErr
		}))
		result, err := service.Grant(ctx, reward)
		if result != wantResult || err != wantErr {
			t.Fatalf("Grant = %+v, %v; result/error not preserved", result, err)
		}
	}
}
