package social

import (
	"context"
	"errors"
	"time"
)

type Actor struct {
	UserID      int
	SessionHash string
}
type SendInput struct {
	Body        string `json:"body"`
	CharacterID int    `json:"character_id"`
}
type DeleteInput struct {
	MessageID int64  `json:"message_id"`
	Reason    string `json:"reason"`
}
type BanInput struct {
	UserID        int    `json:"user_id"`
	Scope         string `json:"scope"`
	DurationHours *int   `json:"duration_hours"`
	Reason        string `json:"reason"`
}
type RevokeInput struct {
	BanID  int64  `json:"ban_id"`
	Reason string `json:"reason"`
}
type Message struct {
	ID          int64      `json:"id"`
	UserID      int        `json:"user_id"`
	CharacterID int        `json:"character_id"`
	AuthorName  string     `json:"author_name"`
	AuthorRole  string     `json:"author_role"`
	Body        *string    `json:"body"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}
type Ban struct {
	ID              int64      `json:"id"`
	UserID          int        `json:"user_id"`
	Scope           string     `json:"scope"`
	Reason          string     `json:"reason"`
	CreatedByUserID int        `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	RevokedAt       *time.Time `json:"revoked_at"`
	RevokedByUserID *int       `json:"revoked_by_user_id"`
}
type Action struct {
	ID          int64     `json:"id"`
	ActorUserID *int      `json:"actor_user_id"`
	UserID      int       `json:"user_id"`
	Scope       string    `json:"scope"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
	BanID       *int64    `json:"ban_id"`
	MessageID   *int64    `json:"message_id"`
	CreatedAt   time.Time `json:"created_at"`
}
type ChatPage struct {
	Messages     []Message `json:"messages"`
	NextBeforeID *int64    `json:"next_before_id"`
	ChatBan      *Ban      `json:"chat_ban"`
}
type BanPage struct {
	Bans         []Ban  `json:"bans"`
	NextBeforeID *int64 `json:"next_before_id"`
}
type ActionPage struct {
	Actions      []Action `json:"actions"`
	NextBeforeID *int64   `json:"next_before_id"`
}

type Moderator struct {
	UserID        int     `json:"user_id"`
	Login         string  `json:"login"`
	Role          string  `json:"role"`
	CharacterName *string `json:"character_name"`
}

type ModeratorList struct {
	Moderators []Moderator `json:"moderators"`
}
type Repository interface {
	Chat(context.Context, Actor, int64) (*ChatPage, error)
	Send(context.Context, Actor, SendInput) (*Message, error)
	Delete(context.Context, Actor, DeleteInput) error
	Bans(context.Context, Actor, int64, int) (*BanPage, error)
	Ban(context.Context, Actor, BanInput) (*Ban, error)
	Revoke(context.Context, Actor, RevokeInput) error
	Actions(context.Context, Actor, int64) (*ActionPage, error)
	Moderators(context.Context, Actor) (*ModeratorList, error)
}

var (
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("not found")
	ErrChatBanned       = errors.New("chat banned")
	ErrNoCharacter      = errors.New("no active character")
	ErrCharacterChanged = errors.New("active character changed")
	ErrTooFast          = errors.New("chat rate limit")
	ErrBanActive        = errors.New("ban already active")
)
