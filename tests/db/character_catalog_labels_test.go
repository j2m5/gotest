package db_test

import (
	"testing"

	"gotest/internal/db"
)

func TestCharacterCatalogUsesDisplayAliasesWithNameFallback(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `UPDATE factions SET alias='  Империя  ' WHERE code='sith_empire';
		UPDATE factions SET alias='' WHERE code='galactic_republic';
		UPDATE classes SET alias='  Сит-воин  ' WHERE code='sith_warrior';
		UPDATE classes SET alias=NULL WHERE code='sith_inquisitor';
		UPDATE classes SET alias='   ' WHERE code='imperial_agent'`)
	catalog, err := db.NewCharacterRepository(s.pool).Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"sith_empire":       "Империя",
		"galactic_republic": "Galactic Republic",
		"sith_warrior":      "Сит-воин",
		"sith_inquisitor":   "Sith Inquisitor",
		"imperial_agent":    "Imperial Agent",
	}
	for _, faction := range catalog.Factions {
		if expected, ok := want[faction.Code]; ok {
			if faction.Name != expected {
				t.Errorf("faction %s label = %q, want %q", faction.Code, faction.Name, expected)
			}
			delete(want, faction.Code)
		}
	}
	for _, class := range catalog.Classes {
		if expected, ok := want[class.Code]; ok {
			if class.Name != expected {
				t.Errorf("class %s label = %q, want %q", class.Code, class.Name, expected)
			}
			delete(want, class.Code)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing catalog entries: %v", want)
	}
}
