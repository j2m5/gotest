package character_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gotest/internal/character"
	"gotest/internal/validation"
)

type repository struct {
	calls  int
	actor  character.Actor
	input  character.CreateInput
	active *int
	err    error
}

func (r *repository) Catalog(context.Context) (*character.Catalog, error) {
	r.calls++
	return &character.Catalog{}, r.err
}
func (r *repository) List(_ context.Context, actor character.Actor) (*character.Roster, error) {
	r.calls++
	r.actor = actor
	return &character.Roster{}, r.err
}
func (r *repository) Profile(_ context.Context, actor character.Actor, id int) (*character.PublicProfile, error) {
	r.calls++
	r.actor = actor
	return &character.PublicProfile{ID: id}, r.err
}
func (r *repository) Create(_ context.Context, actor character.Actor, input character.CreateInput) (*character.Character, error) {
	r.calls++
	r.actor, r.input = actor, input
	return &character.Character{ID: 4, Name: input.Name, ClassID: input.ClassID, FactionID: input.FactionID}, r.err
}
func (r *repository) SetActive(_ context.Context, actor character.Actor, id *int) (*character.Character, error) {
	r.calls++
	r.actor, r.active = actor, id
	if id == nil {
		return nil, r.err
	}
	return &character.Character{ID: *id}, r.err
}
func owner() character.Actor { return character.Actor{UserID: 7, SessionHash: strings.Repeat("a", 64)} }
func createInput(name string) character.CreateInput {
	return character.CreateInput{Name: name, FactionID: 2, ClassID: 8}
}

func TestCreateNormalizesAllowedUnicodeNames(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"  Darth   Malgus  ", "Darth Malgus"}, {"Оби-Ван", "Оби-Ван"}, {"О'Брайен", "О'Брайен"},
		{"Zoe\u0308", "Zoë"}, {strings.Repeat("Я", 32), strings.Repeat("Я", 32)},
	} {
		t.Run(tc.input, func(t *testing.T) {
			repo := &repository{}
			got, err := character.NewService(repo).Create(context.Background(), owner(), createInput(tc.input))
			if err != nil || got.Name != tc.want || repo.input.Name != tc.want || repo.actor != owner() || repo.calls != 1 {
				t.Fatalf("create: got=%+v err=%v repository=%+v", got, err, repo)
			}
		})
	}
}

func TestCreateRejectsInvalidNamesBeforePersistence(t *testing.T) {
	for _, name := range []string{"", "  ", "Ab", strings.Repeat("Я", 33), "Darth1", "---", "O\tBri", "O\nBri", "Имя😀", "漢字名", "<Malgus>", "-Malgus", "Malgus'", "Obi\x00wan", "A\xffbc"} {
		t.Run(name, func(t *testing.T) {
			repo := &repository{}
			_, err := character.NewService(repo).Create(context.Background(), owner(), createInput(name))
			var invalid *validation.Error
			if !errors.As(err, &invalid) || len(invalid.Fields) == 0 || repo.calls != 0 {
				t.Fatalf("invalid name reached repository: err=%v calls=%d", err, repo.calls)
			}
		})
	}
}

func TestInvalidIDsDoNotReachRepository(t *testing.T) {
	repo := &repository{}
	service := character.NewService(repo)
	for _, input := range []character.CreateInput{{Name: "Malgus", ClassID: 0, FactionID: 1}, {Name: "Malgus", ClassID: 1, FactionID: -1}} {
		_, err := service.Create(context.Background(), owner(), input)
		var invalid *validation.Error
		if !errors.As(err, &invalid) {
			t.Fatalf("expected validation, got %v", err)
		}
	}
	for _, id := range []int{0, -1} {
		_, err := service.SetActive(context.Background(), owner(), &id)
		var invalid *validation.Error
		if !errors.As(err, &invalid) {
			t.Fatalf("expected validation, got %v", err)
		}
	}
	if repo.calls != 0 {
		t.Fatal("invalid IDs reached persistence")
	}
}

func TestSelectionAndClearKeepAccountAndSessionIdentity(t *testing.T) {
	repo := &repository{}
	service := character.NewService(repo)
	id := 9
	got, err := service.SetActive(context.Background(), owner(), &id)
	if err != nil || got.ID != id || repo.active == nil || *repo.active != id || repo.actor != owner() {
		t.Fatalf("select: %v %v", got, err)
	}
	got, err = service.SetActive(context.Background(), owner(), nil)
	if err != nil || got != nil || repo.active != nil || repo.actor != owner() {
		t.Fatalf("clear: %v %v", got, err)
	}
	if _, err := service.List(context.Background(), owner()); err != nil || repo.actor != owner() {
		t.Fatalf("list: %v", err)
	}
}

func TestMissingActorAndCancelledContextHaveNoSideEffects(t *testing.T) {
	repo := &repository{}
	service := character.NewService(repo)
	for _, actor := range []character.Actor{{}, {UserID: 7}, {SessionHash: strings.Repeat("a", 64)}} {
		if _, err := service.Create(context.Background(), actor, createInput("Malgus")); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("create: %v", err)
		}
		if _, err := service.SetActive(context.Background(), actor, nil); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("select: %v", err)
		}
		if _, err := service.List(context.Background(), actor); !errors.Is(err, character.ErrUnauthenticated) {
			t.Fatalf("list: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Create(ctx, owner(), createInput("Malgus")); !errors.Is(err, context.Canceled) {
		t.Fatalf("create canceled: %v", err)
	}
	if _, err := service.SetActive(ctx, owner(), nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("select canceled: %v", err)
	}
	if _, err := service.List(ctx, owner()); !errors.Is(err, context.Canceled) {
		t.Fatalf("list canceled: %v", err)
	}
	if _, err := service.Catalog(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("catalog canceled: %v", err)
	}
	if repo.calls != 0 {
		t.Fatal("invalid actor or cancellation reached repository")
	}
}

func TestRepositoryErrorsKeepTheirMeaning(t *testing.T) {
	for _, cause := range []error{character.ErrNameTaken, character.ErrInvalidClass, character.ErrNotFound, character.ErrUnauthenticated, errors.New("database unavailable")} {
		service := character.NewService(&repository{err: cause})
		if _, err := service.Create(context.Background(), owner(), createInput("Malgus")); !errors.Is(err, cause) {
			t.Fatalf("error changed: %v", err)
		}
		if _, err := service.SetActive(context.Background(), owner(), nil); !errors.Is(err, cause) {
			t.Fatalf("error changed: %v", err)
		}
		if _, err := service.List(context.Background(), owner()); !errors.Is(err, cause) {
			t.Fatalf("error changed: %v", err)
		}
		if _, err := service.Catalog(context.Background()); !errors.Is(err, cause) {
			t.Fatalf("error changed: %v", err)
		}
	}
}
