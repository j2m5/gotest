package social

import (
	"context"
	"gotest/internal/auth"
	"gotest/internal/validation"
	"strings"
	"unicode/utf8"
)

type Service struct{ repository Repository }

func NewService(r Repository) *Service { return &Service{r} }

func (s *Service) Moderators(ctx context.Context, a Actor) (*ModeratorList, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	return s.repository.Moderators(ctx, a)
}
func check(ctx context.Context, a Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.UserID <= 0 || a.SessionHash == "" {
		return auth.ErrUnauthenticated
	}
	return nil
}
func invalid(field string) error {
	messages := map[string]string{
		"body":           "Сообщение должно содержать от 1 до 1000 символов",
		"reason":         "Причина должна содержать от 3 до 500 символов",
		"character_id":   "Выберите персонажа",
		"user_id":        "Укажите положительный ID аккаунта",
		"message_id":     "Укажите положительный ID сообщения",
		"ban_id":         "Укажите положительный ID блокировки",
		"before_id":      "Курсор должен быть неотрицательным целым числом",
		"scope":          "Выберите область блокировки: account, chat или forum",
		"duration_hours": "Срок блокировки — целое число часов от 1 до 168; бессрочная блокировка доступна только для аккаунта",
	}
	return &validation.Error{Fields: validation.Fields{field: {messages[field]}}}
}
func text(v string, min, max int) bool {
	return utf8.ValidString(v) && !strings.ContainsRune(v, 0) && utf8.RuneCountInString(v) >= min && utf8.RuneCountInString(v) <= max
}
func (s *Service) Chat(ctx context.Context, a Actor, before int64) (*ChatPage, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	if before < 0 {
		return nil, invalid("before_id")
	}
	return s.repository.Chat(ctx, a, before)
}
func (s *Service) Send(ctx context.Context, a Actor, in SendInput) (*Message, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	in.Body = strings.TrimSpace(in.Body)
	if !text(in.Body, 1, 1000) {
		return nil, invalid("body")
	}
	if in.CharacterID <= 0 {
		return nil, invalid("character_id")
	}
	return s.repository.Send(ctx, a, in)
}
func (s *Service) Delete(ctx context.Context, a Actor, in DeleteInput) error {
	if err := check(ctx, a); err != nil {
		return err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !text(in.Reason, 3, 500) {
		return invalid("reason")
	}
	if in.MessageID <= 0 {
		return invalid("message_id")
	}
	return s.repository.Delete(ctx, a, in)
}
func (s *Service) Ban(ctx context.Context, a Actor, in BanInput) (*Ban, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !text(in.Reason, 3, 500) {
		return nil, invalid("reason")
	}
	if in.UserID <= 0 {
		return nil, invalid("user_id")
	}
	if in.Scope != "chat" && in.Scope != "forum" && in.Scope != "account" {
		return nil, invalid("scope")
	}
	if in.DurationHours == nil {
		if in.Scope != "account" {
			return nil, invalid("duration_hours")
		}
	} else if *in.DurationHours < 1 || *in.DurationHours > 168 {
		return nil, invalid("duration_hours")
	}
	return s.repository.Ban(ctx, a, in)
}
func (s *Service) Revoke(ctx context.Context, a Actor, in RevokeInput) error {
	if err := check(ctx, a); err != nil {
		return err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !text(in.Reason, 3, 500) {
		return invalid("reason")
	}
	if in.BanID <= 0 {
		return invalid("ban_id")
	}
	return s.repository.Revoke(ctx, a, in)
}
func (s *Service) Bans(ctx context.Context, a Actor, before int64, user int) (*BanPage, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	if before < 0 {
		return nil, invalid("before_id")
	}
	if user < 0 {
		return nil, invalid("user_id")
	}
	return s.repository.Bans(ctx, a, before, user)
}
func (s *Service) Actions(ctx context.Context, a Actor, before int64) (*ActionPage, error) {
	if err := check(ctx, a); err != nil {
		return nil, err
	}
	if before < 0 {
		return nil, invalid("before_id")
	}
	return s.repository.Actions(ctx, a, before)
}
