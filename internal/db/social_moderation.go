package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"gotest/internal/auth"
	"gotest/internal/social"
)

func (r *SocialRepository) Ban(ctx context.Context, a social.Actor, in social.BanInput) (*social.Ban, error) {
	tx, users, err := r.begin(ctx, a, in.UserID)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	if !permitted(users[a.UserID].role, users[in.UserID].role, in.Scope, a.UserID == in.UserID) {
		return nil, social.ErrForbidden
	}
	existing, err := activeBan(ctx, tx, in.UserID, in.Scope)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, social.ErrBanActive
	}
	b, err := scanBan(tx.QueryRow(ctx, `WITH stamp AS (SELECT clock_timestamp() AS at) INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,created_at,updated_at,expires_at) SELECT $1,$2,$3,$4,at,at,CASE WHEN $5::integer IS NULL THEN NULL ELSE at+($5*interval '1 hour') END FROM stamp RETURNING `+banColumns, in.UserID, in.Scope, in.Reason, a.UserID, in.DurationHours))
	if err != nil {
		return nil, err
	}
	if in.Scope == "account" {
		if _, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, in.UserID); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `UPDATE users SET character_id=NULL,updated_at=clock_timestamp() WHERE id=$1`, in.UserID); err != nil {
			return nil, err
		}
	}
	if err = audit(ctx, tx, a, in.UserID, in.Scope, "ban", in.Reason, &b.ID, nil); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return b, nil
}
func (r *SocialRepository) Revoke(ctx context.Context, a social.Actor, in social.RevokeInput) error {
	var target int
	err := r.pool.QueryRow(ctx, `SELECT user_id FROM user_bans WHERE id=$1`, in.BanID).Scan(&target)
	if errors.Is(err, pgx.ErrNoRows) {
		return social.ErrNotFound
	}
	if err != nil {
		return err
	}
	tx, users, err := r.begin(ctx, a, target)
	if err != nil {
		return err
	}
	defer rollbackSocialTx(tx)
	b, err := scanBan(tx.QueryRow(ctx, `SELECT `+banColumns+` FROM user_bans WHERE id=$1 FOR UPDATE`, in.BanID))
	if err != nil {
		return err
	}
	if !permitted(users[a.UserID].role, users[target].role, b.Scope, a.UserID == target) {
		return social.ErrForbidden
	}
	if b.RevokedAt != nil {
		return social.ErrNotFound
	}
	if _, err = tx.Exec(ctx, `UPDATE user_bans SET revoked_at=clock_timestamp(),revoked_by_user_id=$2,updated_at=clock_timestamp() WHERE id=$1`, in.BanID, a.UserID); err != nil {
		return err
	}
	if err = audit(ctx, tx, a, target, b.Scope, "revoke", in.Reason, &in.BanID, nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *SocialRepository) Bans(ctx context.Context, a social.Actor, before int64, user int) (*social.BanPage, error) {
	tx, users, err := r.begin(ctx, a, 0)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	if users[a.UserID].role != auth.RoleModerator && users[a.UserID].role != auth.RoleAdministrator {
		return nil, social.ErrForbidden
	}
	p := &social.BanPage{Bans: []social.Ban{}}
	rows, err := tx.Query(ctx, `SELECT `+banColumns+` FROM user_bans WHERE ($1::bigint=0 OR id<$1) AND ($2::integer=0 OR user_id=$2) ORDER BY id DESC LIMIT 51`, before, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		b, e := scanBan(rows)
		if e != nil {
			return nil, e
		}
		p.Bans = append(p.Bans, *b)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(p.Bans) > 50 {
		p.Bans = p.Bans[:50]
		id := p.Bans[49].ID
		p.NextBeforeID = &id
	}
	return p, tx.Commit(ctx)
}
func (r *SocialRepository) Actions(ctx context.Context, a social.Actor, before int64) (*social.ActionPage, error) {
	tx, users, err := r.begin(ctx, a, 0)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	if users[a.UserID].role != auth.RoleModerator && users[a.UserID].role != auth.RoleAdministrator {
		return nil, social.ErrForbidden
	}
	p := &social.ActionPage{Actions: []social.Action{}}
	rows, err := tx.Query(ctx, `SELECT id,actor_user_id,user_id,scope,action,reason,ban_id,message_id,created_at FROM moderation_actions WHERE ($1::bigint=0 OR id<$1) ORDER BY id DESC LIMIT 51`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item social.Action
		if err = rows.Scan(&item.ID, &item.ActorUserID, &item.UserID, &item.Scope, &item.Action, &item.Reason, &item.BanID, &item.MessageID, &item.CreatedAt); err != nil {
			return nil, err
		}
		p.Actions = append(p.Actions, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(p.Actions) > 50 {
		p.Actions = p.Actions[:50]
		id := p.Actions[49].ID
		p.NextBeforeID = &id
	}
	return p, tx.Commit(ctx)
}
