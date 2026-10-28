package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/social"
	"sort"
	"time"
)

type SocialRepository struct{ pool *pgxpool.Pool }

var _ social.Repository = (*SocialRepository)(nil)

func rollbackSocialTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}

func NewSocialRepository(p *pgxpool.Pool) *SocialRepository { return &SocialRepository{p} }

type socialUser struct {
	role      int16
	character *int
}

// All social operations lock accounts first, in ascending order. Session and
// role checks happen after the locks so a waiter cannot retain revoked rights.
func (r *SocialRepository) begin(ctx context.Context, a social.Actor, target int) (pgx.Tx, map[int]socialUser, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	fail := func(e error) (pgx.Tx, map[int]socialUser, error) { rollbackSocialTx(tx); return nil, nil, e }
	ids := []int{a.UserID}
	if target > 0 && target != a.UserID {
		ids = append(ids, target)
	}
	sort.Ints(ids)
	users := map[int]socialUser{}
	for _, id := range ids {
		var u socialUser
		err = tx.QueryRow(ctx, `SELECT role_id,character_id FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&u.role, &u.character)
		if errors.Is(err, pgx.ErrNoRows) {
			if id == a.UserID {
				return fail(auth.ErrUnauthenticated)
			}
			return fail(social.ErrNotFound)
		}
		if err != nil {
			return fail(err)
		}
		users[id] = u
	}
	var ok bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token=$1 AND s.user_id=$2 AND s.expires_at>clock_timestamp() AND u.email_verified_at IS NOT NULL AND NOT EXISTS(SELECT 1 FROM user_bans b WHERE b.user_id=u.id AND b.scope='account' AND b.revoked_at IS NULL AND (b.expires_at IS NULL OR b.expires_at>clock_timestamp())))`, a.SessionHash, a.UserID).Scan(&ok)
	if err != nil {
		return fail(err)
	}
	if !ok {
		return fail(auth.ErrUnauthenticated)
	}
	return tx, users, nil
}
func permitted(actor, target int16, scope string, self bool) bool {
	if self {
		return false
	}
	if actor == auth.RoleAdministrator {
		return true
	}
	return actor == auth.RoleModerator && target == auth.RoleUser && scope != "account"
}

const banColumns = `id,user_id,scope,reason,created_by_user_id,created_at,expires_at,revoked_at,revoked_by_user_id`

func scanBan(row pgx.Row) (*social.Ban, error) {
	b := &social.Ban{}
	err := row.Scan(&b.ID, &b.UserID, &b.Scope, &b.Reason, &b.CreatedByUserID, &b.CreatedAt, &b.ExpiresAt, &b.RevokedAt, &b.RevokedByUserID)
	return b, err
}
func activeBan(ctx context.Context, tx pgx.Tx, id int, scope string) (*social.Ban, error) {
	b, err := scanBan(tx.QueryRow(ctx, `SELECT `+banColumns+` FROM user_bans WHERE user_id=$1 AND scope=$2 AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at>clock_timestamp()) ORDER BY id DESC LIMIT 1`, id, scope))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return b, err
}
func audit(ctx context.Context, tx pgx.Tx, a social.Actor, user int, scope, action, reason string, ban, message *int64) error {
	_, err := tx.Exec(ctx, `INSERT INTO moderation_actions(actor_user_id,user_id,scope,action,reason,ban_id,message_id) VALUES($1,$2,$3,$4,$5,$6,$7)`, a.UserID, user, scope, action, reason, ban, message)
	return err
}

const messageColumns = `id,user_id,character_id,author_name,author_role,CASE WHEN deleted_at IS NULL THEN body END,created_at,deleted_at`

func scanMessage(row pgx.Row) (*social.Message, error) {
	m := &social.Message{}
	err := row.Scan(&m.ID, &m.UserID, &m.CharacterID, &m.AuthorName, &m.AuthorRole, &m.Body, &m.CreatedAt, &m.DeletedAt)
	return m, err
}
func (r *SocialRepository) Send(ctx context.Context, a social.Actor, in social.SendInput) (*social.Message, error) {
	tx, users, err := r.begin(ctx, a, 0)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	b, err := activeBan(ctx, tx, a.UserID, "chat")
	if err != nil {
		return nil, err
	}
	if b != nil {
		return nil, social.ErrChatBanned
	}
	u := users[a.UserID]
	if u.character == nil {
		return nil, social.ErrNoCharacter
	}
	if *u.character != in.CharacterID {
		return nil, social.ErrCharacterChanged
	}
	var name string
	err = tx.QueryRow(ctx, `SELECT name FROM characters WHERE id=$1 AND user_id=$2`, *u.character, a.UserID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, social.ErrNoCharacter
	}
	if err != nil {
		return nil, err
	}
	var fast bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM chat_messages WHERE user_id=$1 AND created_at>clock_timestamp()-interval '2 seconds')`, a.UserID).Scan(&fast)
	if err != nil {
		return nil, err
	}
	if fast {
		return nil, social.ErrTooFast
	}
	m, err := scanMessage(tx.QueryRow(ctx, `INSERT INTO chat_messages(user_id,character_id,author_name,author_role,body) VALUES($1,$2,$3,$4,$5) RETURNING `+messageColumns, a.UserID, *u.character, name, auth.RoleName(u.role), in.Body))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return m, nil
}
func (r *SocialRepository) Chat(ctx context.Context, a social.Actor, before int64) (*social.ChatPage, error) {
	tx, _, err := r.begin(ctx, a, 0)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	p := &social.ChatPage{Messages: []social.Message{}}
	p.ChatBan, err = activeBan(ctx, tx, a.UserID, "chat")
	if err != nil {
		return nil, err
	}
	// Names remain the original character snapshot. Role badges/colors reflect
	// the account's current authority, including after promotion or demotion.
	rows, err := tx.Query(ctx, `SELECT m.id,m.user_id,m.character_id,m.author_name,role.name,
		CASE WHEN m.deleted_at IS NULL THEN m.body END,m.created_at,m.deleted_at
		FROM chat_messages m JOIN users u ON u.id=m.user_id JOIN account_roles role ON role.id=u.role_id
		WHERE ($1::bigint=0 OR m.id<$1) ORDER BY m.id DESC LIMIT 51`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		m, e := scanMessage(rows)
		if e != nil {
			return nil, e
		}
		p.Messages = append(p.Messages, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(p.Messages) > 50 {
		p.Messages = p.Messages[:50]
		id := p.Messages[49].ID
		p.NextBeforeID = &id
	}
	return p, tx.Commit(ctx)
}
func (r *SocialRepository) Delete(ctx context.Context, a social.Actor, in social.DeleteInput) error {
	var user int
	err := r.pool.QueryRow(ctx, `SELECT user_id FROM chat_messages WHERE id=$1`, in.MessageID).Scan(&user)
	if errors.Is(err, pgx.ErrNoRows) {
		return social.ErrNotFound
	}
	if err != nil {
		return err
	}
	tx, users, err := r.begin(ctx, a, user)
	if err != nil {
		return err
	}
	defer rollbackSocialTx(tx)
	role := users[a.UserID].role
	if !(role == auth.RoleAdministrator || (role == auth.RoleModerator && (users[user].role == auth.RoleUser || a.UserID == user))) {
		return social.ErrForbidden
	}
	tag, err := tx.Exec(ctx, `UPDATE chat_messages SET deleted_at=clock_timestamp(),deleted_by_user_id=$2,updated_at=clock_timestamp() WHERE id=$1 AND deleted_at IS NULL`, in.MessageID, a.UserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return social.ErrNotFound
	}
	if err = audit(ctx, tx, a, user, "chat", "delete_message", in.Reason, nil, &in.MessageID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
