package character_test

import (
	"context"
	"errors"
	"testing"

	"gotest/internal/character"
	"gotest/internal/validation"
)

func TestCharacterValidationNamesAffectedFields(t *testing.T) {
	repo := &repository{}
	service := character.NewService(repo)
	_, err := service.Create(context.Background(), owner(), character.CreateInput{Name: "-"})
	var invalid *validation.Error
	if !errors.As(err, &invalid) {
		t.Fatalf("expected field validation, got %v", err)
	}
	if len(invalid.Fields) != 3 {
		t.Fatalf("unexpected fields: %v", invalid.Fields)
	}
	for _, field := range []string{"name", "faction_id", "class_id"} {
		if len(invalid.Fields[field]) != 1 || invalid.Fields[field][0] == "" {
			t.Errorf("missing error for %s", field)
		}
	}
	id := 0
	_, err = service.SetActive(context.Background(), owner(), &id)
	if !errors.As(err, &invalid) || len(invalid.Fields["character_id"]) != 1 {
		t.Fatalf("selection field validation: %v", err)
	}
	if repo.calls != 0 {
		t.Fatal("invalid fields reached persistence")
	}
}
