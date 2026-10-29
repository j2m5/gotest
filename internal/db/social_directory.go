package db

import (
	"context"

	"gotest/internal/social"
)

// Moderators is a player-visible directory, separate from private moderation
// records. One row represents one account, regardless of its character count.
func (r *SocialRepository) Moderators(ctx context.Context, a social.Actor) (*social.ModeratorList, error) {
	tx, _, err := r.begin(ctx, a, 0)
	if err != nil {
		return nil, err
	}
	defer rollbackSocialTx(tx)
	list := &social.ModeratorList{Moderators: []social.Moderator{}}
	rows, err := tx.Query(ctx, `SELECT u.id,u.login,role.name,c.id,c.name
		FROM users u JOIN account_roles role ON role.id=u.role_id
		LEFT JOIN characters c ON c.id=u.character_id AND c.user_id=u.id
		WHERE u.role_id=2 AND u.email_verified_at IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM user_bans b WHERE b.user_id=u.id AND b.scope='account'
			AND b.revoked_at IS NULL AND (b.expires_at IS NULL OR b.expires_at>clock_timestamp()))
		ORDER BY lower(u.login),u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var member social.Moderator
		if err = rows.Scan(&member.UserID, &member.Login, &member.Role, &member.CharacterID, &member.CharacterName); err != nil {
			return nil, err
		}
		list.Moderators = append(list.Moderators, member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, tx.Commit(ctx)
}
