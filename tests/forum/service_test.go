package forum_test

import (
	"context"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/forum"
	"gotest/internal/validation"
	"strings"
	"testing"
)

// Rejecting malformed input before storage protects every transport using the service.
func TestForumRejectsMalformedWritesAndCursors(t *testing.T) {
	s := forum.NewService(nil)
	ctx := context.Background()
	a := forum.Actor{UserID: 1, SessionHash: "session"}
	cases := []struct {
		name string
		run  func() error
	}{
		{"short title", func() error {
			_, e := s.Create(ctx, a, forum.CreateInput{CategoryID: 1, CharacterID: 1, Title: "ab", Body: "ok"})
			return e
		}},
		{"long Unicode title", func() error {
			_, e := s.Create(ctx, a, forum.CreateInput{CategoryID: 1, CharacterID: 1, Title: strings.Repeat("я", 121), Body: "ok"})
			return e
		}},
		{"empty body", func() error {
			_, e := s.Reply(ctx, a, forum.ReplyInput{TopicID: 1, CharacterID: 1, Body: " \t"})
			return e
		}},
		{"invalid UTF8", func() error {
			_, e := s.Reply(ctx, a, forum.ReplyInput{TopicID: 1, CharacterID: 1, Body: string([]byte{255})})
			return e
		}},
		{"NUL", func() error {
			_, e := s.Reply(ctx, a, forum.ReplyInput{TopicID: 1, CharacterID: 1, Body: "a\x00b"})
			return e
		}},
		{"long body", func() error {
			_, e := s.Reply(ctx, a, forum.ReplyInput{TopicID: 1, CharacterID: 1, Body: strings.Repeat("я", 10001)})
			return e
		}},
		{"missing character", func() error { _, e := s.Reply(ctx, a, forum.ReplyInput{TopicID: 1, Body: "ok"}); return e }},
		{"version", func() error {
			_, e := s.EditPost(ctx, a, forum.EditPostInput{PostID: 1, CharacterID: 1, Body: "ok"})
			return e
		}},
		{"topic id", func() error { _, e := s.Topic(ctx, a, 0, 0); return e }},
		{"after id", func() error { _, e := s.Topic(ctx, a, 1, -1); return e }},
		{"category id", func() error { _, e := s.Topics(ctx, a, forum.TopicsInput{}); return e }},
		{"post action", func() error {
			return s.Moderate(ctx, a, forum.ModerateInput{Target: "post", TargetID: 1, Action: "close", Reason: "reason"})
		}},
		{"reason", func() error {
			return s.Moderate(ctx, a, forum.ModerateInput{Target: "topic", TargetID: 1, Action: "hide", Reason: " x "})
		}},
	}
	for _, v := range []string{"p:0", "n:-1", "n:01", "p:+1", "x:1", "n:9223372036854775808", "p:1:2", "  n:2"} {
		cursor := v
		cases = append(cases, struct {
			name string
			run  func() error
		}{"cursor " + v, func() error { _, e := s.Topics(ctx, a, forum.TopicsInput{CategoryID: 1, Cursor: cursor}); return e }})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ve *validation.Error
			if err := tc.run(); !errors.As(err, &ve) {
				t.Fatalf("validation=%v", err)
			}
		})
	}
	if _, e := s.Categories(ctx, forum.Actor{}); !errors.Is(e, auth.ErrUnauthenticated) {
		t.Fatalf("auth=%v", e)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := s.Categories(cancelled, a); !errors.Is(e, context.Canceled) {
		t.Fatalf("context=%v", e)
	}
}

type captureRepo struct {
	forum.Repository
	create   forum.CreateInput
	edit     forum.EditTopicInput
	moderate forum.ModerateInput
}

func (r *captureRepo) Create(_ context.Context, _ forum.Actor, in forum.CreateInput) (*forum.Topic, error) {
	r.create = in
	return &forum.Topic{Body: &in.Body}, nil
}
func (r *captureRepo) EditTopic(_ context.Context, _ forum.Actor, in forum.EditTopicInput) (*forum.Topic, error) {
	r.edit = in
	return &forum.Topic{Body: &in.Body}, nil
}
func (r *captureRepo) Moderate(_ context.Context, _ forum.Actor, in forum.ModerateInput) error {
	r.moderate = in
	return nil
}
func TestForumTrimsTextAndCountsUnicodeCharacters(t *testing.T) {
	r := &captureRepo{}
	s := forum.NewService(r)
	a := forum.Actor{UserID: 1, SessionHash: "s"}
	ctx := context.Background()
	topic, e := s.Create(ctx, a, forum.CreateInput{CategoryID: 2, CharacterID: 1, Title: "  яяя  ", Body: " \n<body>literal</body> \t"})
	if e != nil || topic.Body == nil || *topic.Body != "<body>literal</body>" || r.create.Title != "яяя" {
		t.Fatalf("create=%+v %v", topic, e)
	}
	_, e = s.EditTopic(ctx, a, forum.EditTopicInput{TopicID: 1, CharacterID: 2, Title: strings.Repeat("я", 120), Body: " body ", Version: 1})
	if e != nil || r.edit.Body != "body" {
		t.Fatalf("edit=%+v %v", r.edit, e)
	}
	if e = s.Moderate(ctx, a, forum.ModerateInput{Target: "topic", TargetID: 1, Action: "pin", Reason: " reason "}); e != nil || r.moderate.Reason != "reason" {
		t.Fatalf("moderate=%+v %v", r.moderate, e)
	}
}
