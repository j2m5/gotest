package db_test

import (
	"reflect"
	"testing"

	"gotest/internal/db"
)

func TestCharacterRepositoryPreservesNullableFields(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	repository := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 1, "nullable-fields")
	for _, tc := range []struct {
		name                      string
		description, icon, avatar *string
	}{
		{name: "null"},
		{name: "empty", description: new(""), icon: new(""), avatar: new("")},
		{name: "present", description: new("Catalog description"), icon: new("icon.png"), avatar: new("avatar.png")},
		{name: "cleared"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustExec(t, ctx, s, "UPDATE factions SET description=$1,icon=$2", tc.description, tc.icon)
			mustExec(t, ctx, s, "UPDATE classes SET description=$1,icon=$2", tc.description, tc.icon)
			mustExec(t, ctx, s, "UPDATE characters SET avatar=$1 WHERE id=1", tc.avatar)

			catalog, err := repository.Catalog(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(catalog.Factions) != 2 || len(catalog.Classes) != 8 {
				t.Fatalf("catalog entries lost: %+v", catalog)
			}
			for _, faction := range catalog.Factions {
				if !reflect.DeepEqual(faction.Description, tc.description) || !reflect.DeepEqual(faction.Icon, tc.icon) {
					t.Fatalf("faction %q lost nullable values: %+v", faction.Code, faction)
				}
			}
			for _, class := range catalog.Classes {
				if !reflect.DeepEqual(class.Description, tc.description) || !reflect.DeepEqual(class.Icon, tc.icon) {
					t.Fatalf("class %q lost nullable values: %+v", class.Code, class)
				}
			}

			roster, err := repository.List(ctx, actor)
			if err != nil {
				t.Fatal(err)
			}
			if len(roster.Characters) != 2 || roster.Characters[0].ID != 1 {
				t.Fatalf("unexpected roster: %+v", roster)
			}
			if !reflect.DeepEqual(roster.Characters[0].Avatar, tc.avatar) {
				t.Fatalf("list lost nullable avatar: %+v", roster.Characters[0])
			}
			selected, err := repository.SetActive(ctx, actor, new(1))
			if err != nil || selected == nil || selected.ID != 1 || !reflect.DeepEqual(selected.Avatar, tc.avatar) {
				t.Fatalf("selection lost nullable avatar: %+v, err=%v", selected, err)
			}
		})
	}
}
