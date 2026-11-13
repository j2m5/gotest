package combat_test

import (
	"errors"
	"strings"
	"testing"

	"gotest/internal/combat"
)

func TestCombatStartValidation(t *testing.T) {
	valid := combat.StartInput{CharacterID: 1, LocationID: 2, NPCSpawnID: 3, Key: "fight:one/attempt_1"}
	for _, key := range []string{"a", valid.Key, strings.Repeat("Z", 128)} {
		in := valid
		in.Key = key
		if err := combat.ValidateStart(in, combat.NPCStats{HP: 100}); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range []string{"", " ", ":one", "test\n", "test\r", "бой", strings.Repeat("a", 129)} {
		in := valid
		in.Key = key
		if err := combat.ValidateStart(in, combat.NPCStats{HP: 100}); !errors.Is(err, combat.ErrInvalidInput) {
			t.Fatalf("key %q: %v", key, err)
		}
	}
	for _, in := range []combat.StartInput{
		{LocationID: 2, NPCSpawnID: 3, Key: "a"},
		{CharacterID: 1, NPCSpawnID: 3, Key: "a"},
		{CharacterID: 1, LocationID: 2, NPCSpawnID: -1, Key: "a"},
	} {
		if err := combat.ValidateStart(in, combat.NPCStats{HP: 100}); !errors.Is(err, combat.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	for _, hp := range []int{-1, 0, 1 << 31} {
		if err := combat.ValidateStart(valid, combat.NPCStats{HP: hp}); !errors.Is(err, combat.ErrInvalidInput) {
			t.Fatalf("HP %d: %v", hp, err)
		}
	}
}
