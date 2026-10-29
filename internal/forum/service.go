package forum

import (
	"context"
	"gotest/internal/auth"
	"gotest/internal/validation"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Service struct{ repository Repository }

func NewService(r Repository) *Service { return &Service{r} }
func check(ctx context.Context, a Actor) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if a.UserID <= 0 || a.SessionHash == "" {
		return auth.ErrUnauthenticated
	}
	return nil
}
func invalid(field string) error {
	return &validation.Error{Fields: validation.Fields{field: {"Недопустимое значение"}}}
}
func text(v string, min, max int) bool {
	return utf8.ValidString(v) && !strings.ContainsRune(v, 0) && utf8.RuneCountInString(v) >= min && utf8.RuneCountInString(v) <= max
}

// ParseCursor accepts only canonical positive int64 identifiers and a pin bucket.
func ParseCursor(v string) (bool, int64, error) {
	if v == "" {
		return false, 0, nil
	}
	if len(v) < 3 || (v[:2] != "p:" && v[:2] != "n:") {
		return false, 0, invalid("cursor")
	}
	id, e := strconv.ParseInt(v[2:], 10, 64)
	if e != nil || id <= 0 || strconv.FormatInt(id, 10) != v[2:] {
		return false, 0, invalid("cursor")
	}
	return v[:2] == "p:", id, nil
}
func (s *Service) Categories(ctx context.Context, a Actor) (*CategoryList, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	return s.repository.Categories(ctx, a)
}
func (s *Service) Topics(ctx context.Context, a Actor, in TopicsInput) (*TopicList, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	if in.CategoryID <= 0 {
		return nil, invalid("category_id")
	}
	if _, _, e := ParseCursor(in.Cursor); e != nil {
		return nil, e
	}
	return s.repository.Topics(ctx, a, in)
}
func (s *Service) Topic(ctx context.Context, a Actor, id, after int64) (*TopicPage, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	if id <= 0 {
		return nil, invalid("topic_id")
	}
	if after < 0 {
		return nil, invalid("after_id")
	}
	return s.repository.Topic(ctx, a, id, after)
}
func (s *Service) Create(ctx context.Context, a Actor, in CreateInput) (*Topic, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	if in.CategoryID <= 0 {
		return nil, invalid("category_id")
	}
	if in.CharacterID <= 0 {
		return nil, invalid("character_id")
	}
	if !text(in.Title, 3, 120) {
		return nil, invalid("title")
	}
	if !text(in.Body, 1, 10000) {
		return nil, invalid("body")
	}
	return s.repository.Create(ctx, a, in)
}
func (s *Service) Reply(ctx context.Context, a Actor, in ReplyInput) (*Post, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	in.Body = strings.TrimSpace(in.Body)
	if in.TopicID <= 0 {
		return nil, invalid("topic_id")
	}
	if in.CharacterID <= 0 {
		return nil, invalid("character_id")
	}
	if !text(in.Body, 1, 10000) {
		return nil, invalid("body")
	}
	return s.repository.Reply(ctx, a, in)
}
func (s *Service) EditTopic(ctx context.Context, a Actor, in EditTopicInput) (*Topic, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	if in.TopicID <= 0 {
		return nil, invalid("topic_id")
	}
	if in.CharacterID <= 0 {
		return nil, invalid("character_id")
	}
	if in.Version <= 0 {
		return nil, invalid("version")
	}
	if !text(in.Title, 3, 120) {
		return nil, invalid("title")
	}
	if !text(in.Body, 1, 10000) {
		return nil, invalid("body")
	}
	return s.repository.EditTopic(ctx, a, in)
}
func (s *Service) EditPost(ctx context.Context, a Actor, in EditPostInput) (*Post, error) {
	if e := check(ctx, a); e != nil {
		return nil, e
	}
	in.Body = strings.TrimSpace(in.Body)
	if in.PostID <= 0 {
		return nil, invalid("post_id")
	}
	if in.CharacterID <= 0 {
		return nil, invalid("character_id")
	}
	if in.Version <= 0 {
		return nil, invalid("version")
	}
	if !text(in.Body, 1, 10000) {
		return nil, invalid("body")
	}
	return s.repository.EditPost(ctx, a, in)
}
func (s *Service) Moderate(ctx context.Context, a Actor, in ModerateInput) error {
	if e := check(ctx, a); e != nil {
		return e
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if in.TargetID <= 0 {
		return invalid("target_id")
	}
	if in.Target != "topic" && in.Target != "post" {
		return invalid("target")
	}
	if in.Action != "hide" && in.Action != "restore" && (in.Target != "topic" || (in.Action != "close" && in.Action != "open" && in.Action != "pin" && in.Action != "unpin")) {
		return invalid("action")
	}
	if !text(in.Reason, 3, 500) {
		return invalid("reason")
	}
	return s.repository.Moderate(ctx, a, in)
}
