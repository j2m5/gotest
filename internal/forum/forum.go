package forum

import (
	"context"
	"errors"
	"gotest/internal/social"
	"time"
)

type Actor = social.Actor
type Category struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreateRole  string `json:"create_role"`
}
type CategoryList struct {
	Categories []Category `json:"categories"`
}
type TopicSummary struct {
	ID          int64      `json:"id"`
	CategoryID  int        `json:"category_id"`
	UserID      int        `json:"user_id"`
	CharacterID int        `json:"character_id"`
	AuthorName  string     `json:"author_name"`
	AuthorRole  string     `json:"author_role"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	PinnedAt    *time.Time `json:"pinned_at"`
	HiddenAt    *time.Time `json:"hidden_at"`
	ReplyCount  int        `json:"reply_count"`
}
type Topic struct {
	TopicSummary
	Body     *string    `json:"body"`
	EditedAt *time.Time `json:"edited_at"`
	Version  int        `json:"version"`
}
type Post struct {
	ID          int64      `json:"id"`
	TopicID     int64      `json:"topic_id"`
	UserID      int        `json:"user_id"`
	CharacterID int        `json:"character_id"`
	AuthorName  string     `json:"author_name"`
	AuthorRole  string     `json:"author_role"`
	Body        *string    `json:"body"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	EditedAt    *time.Time `json:"edited_at"`
	HiddenAt    *time.Time `json:"hidden_at"`
	Version     int        `json:"version"`
}
type TopicList struct {
	Category   Category       `json:"category"`
	Topics     []TopicSummary `json:"topics"`
	NextCursor *string        `json:"next_cursor"`
	ForumBan   *social.Ban    `json:"forum_ban"`
}
type TopicPage struct {
	Topic       Topic       `json:"topic"`
	Posts       []Post      `json:"posts"`
	NextAfterID *int64      `json:"next_after_id"`
	ForumBan    *social.Ban `json:"forum_ban"`
}
type TopicsInput struct {
	CategoryID    int    `json:"category_id"`
	Cursor        string `json:"cursor"`
	IncludeHidden bool   `json:"include_hidden"`
}
type CreateInput struct {
	CategoryID  int    `json:"category_id"`
	CharacterID int    `json:"character_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
}
type ReplyInput struct {
	TopicID     int64  `json:"topic_id"`
	CharacterID int    `json:"character_id"`
	Body        string `json:"body"`
}
type EditTopicInput struct {
	TopicID     int64  `json:"topic_id"`
	CharacterID int    `json:"character_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Version     int    `json:"version"`
}
type EditPostInput struct {
	PostID      int64  `json:"post_id"`
	CharacterID int    `json:"character_id"`
	Body        string `json:"body"`
	Version     int    `json:"version"`
}
type ModerateInput struct {
	Target   string `json:"target"`
	TargetID int64  `json:"target_id"`
	Action   string `json:"action"`
	Reason   string `json:"reason"`
}
type Repository interface {
	Categories(context.Context, Actor) (*CategoryList, error)
	Topics(context.Context, Actor, TopicsInput) (*TopicList, error)
	Topic(context.Context, Actor, int64, int64) (*TopicPage, error)
	Create(context.Context, Actor, CreateInput) (*Topic, error)
	Reply(context.Context, Actor, ReplyInput) (*Post, error)
	EditTopic(context.Context, Actor, EditTopicInput) (*Topic, error)
	EditPost(context.Context, Actor, EditPostInput) (*Post, error)
	Moderate(context.Context, Actor, ModerateInput) error
}

var (
	ErrNotFound = errors.New("forum not found")
	ErrBanned   = errors.New("forum banned")
	ErrClosed   = errors.New("forum closed")
	ErrConflict = errors.New("forum conflict")
	ErrTooFast  = errors.New("forum rate limit")
)
