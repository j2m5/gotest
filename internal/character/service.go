package character

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Catalog(ctx context.Context) (*Catalog, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.repository.Catalog(ctx)
}

func (s *Service) List(ctx context.Context, actor Actor) (*Roster, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, actor)
}

func (s *Service) Create(ctx context.Context, actor Actor, input CreateInput) (*Character, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	input.Name = norm.NFC.String(strings.TrimSpace(input.Name))
	// Only ordinary spaces are collapsed. Tabs and other embedded whitespace
	// remain invalid rather than silently changing the requested name.
	parts := strings.Split(input.Name, " ")
	words := parts[:0]
	for _, part := range parts {
		if part != "" {
			words = append(words, part)
		}
	}
	input.Name = strings.Join(words, " ")
	var messages []string
	if !validName(input.Name) {
		messages = append(messages, "Имя должно содержать 3–32 символа: латинские или кириллические буквы, пробел, дефис или апостроф; начинаться и заканчиваться буквой")
	}
	if input.FactionID <= 0 {
		messages = append(messages, "Выберите фракцию")
	}
	if input.ClassID <= 0 {
		messages = append(messages, "Выберите класс")
	}
	if len(messages) > 0 {
		return nil, &ValidationError{Messages: messages}
	}
	return s.repository.Create(ctx, actor, input)
}

func (s *Service) SetActive(ctx context.Context, actor Actor, id *int) (*Character, error) {
	if err := checkActor(ctx, actor); err != nil {
		return nil, err
	}
	if id != nil && *id <= 0 {
		return nil, &ValidationError{Messages: []string{"Выберите персонажа"}}
	}
	return s.repository.SetActive(ctx, actor, id)
}

func checkActor(ctx context.Context, actor Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if actor.UserID <= 0 || actor.SessionHash == "" {
		return ErrUnauthenticated
	}
	return nil
}

func nameLetter(r rune) bool {
	return unicode.IsLetter(r) && (unicode.Is(unicode.Latin, r) || unicode.Is(unicode.Cyrillic, r))
}

func validName(name string) bool {
	if !utf8.ValidString(name) {
		return false
	}
	runes := []rune(name)
	if len(runes) < 3 || len(runes) > 32 || !nameLetter(runes[0]) || !nameLetter(runes[len(runes)-1]) {
		return false
	}
	for _, r := range runes {
		if !nameLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}
