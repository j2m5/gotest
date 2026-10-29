package db

import (
	"context"
	"gotest/internal/auth"
	"gotest/internal/forum"
	"gotest/internal/social"
)

func (r *ForumRepository) EditTopic(ctx context.Context, a forum.Actor, in forum.EditTopicInput) (*forum.Topic, error) {
	// Ownership is immutable. Resolve it before acquiring any locks so every
	// mutation follows accounts ascending, then topic, then post.
	var user int
	e := r.shared.pool.QueryRow(ctx, `SELECT user_id FROM forum_topics WHERE id=$1`, in.TopicID).Scan(&user)
	if e != nil {
		return nil, forumNotFound(e)
	}
	tx, users, e := r.shared.begin(ctx, a, user)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	topic, e := loadForumTopic(ctx, tx, in.TopicID, true)
	if e != nil {
		return nil, e
	}
	if e = checkSocialSession(ctx, tx, a); e != nil {
		return nil, e
	}
	if e = forumWritableTopic(topic, users[a.UserID].role); e != nil {
		return nil, e
	}
	if user != a.UserID {
		return nil, social.ErrForbidden
	}
	if _, e = forumCharacter(ctx, tx, a, users[a.UserID], in.CharacterID); e != nil {
		return nil, e
	}
	tag, e := tx.Exec(ctx, `UPDATE forum_topics SET title=$2,body=$3,version=version+1,edited_at=clock_timestamp() WHERE id=$1 AND version=$4`, in.TopicID, in.Title, in.Body, in.Version)
	if e != nil {
		return nil, e
	}
	if tag.RowsAffected() == 0 {
		return nil, forum.ErrConflict
	}
	out, e := loadForumTopic(ctx, tx, in.TopicID, false)
	if e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (r *ForumRepository) EditPost(ctx context.Context, a forum.Actor, in forum.EditPostInput) (*forum.Post, error) {
	var user int
	var topicID int64
	e := r.shared.pool.QueryRow(ctx, `SELECT user_id,topic_id FROM forum_posts WHERE id=$1`, in.PostID).Scan(&user, &topicID)
	if e != nil {
		return nil, forumNotFound(e)
	}
	tx, users, e := r.shared.begin(ctx, a, user)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	topic, e := loadForumTopic(ctx, tx, topicID, true)
	if e != nil {
		return nil, e
	}
	if e = forumWritableTopic(topic, users[a.UserID].role); e != nil {
		return nil, e
	}
	post, e := loadForumPost(ctx, tx, in.PostID, true)
	if e != nil {
		return nil, e
	}
	if e = checkSocialSession(ctx, tx, a); e != nil {
		return nil, e
	}
	if post.HiddenAt != nil {
		return nil, forum.ErrNotFound
	}
	if user != a.UserID {
		return nil, social.ErrForbidden
	}
	if _, e = forumCharacter(ctx, tx, a, users[a.UserID], in.CharacterID); e != nil {
		return nil, e
	}
	tag, e := tx.Exec(ctx, `UPDATE forum_posts SET body=$2,version=version+1,edited_at=clock_timestamp() WHERE id=$1 AND version=$3`, in.PostID, in.Body, in.Version)
	if e != nil {
		return nil, e
	}
	if tag.RowsAffected() == 0 {
		return nil, forum.ErrConflict
	}
	out, e := loadForumPost(ctx, tx, in.PostID, false)
	if e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (r *ForumRepository) Moderate(ctx context.Context, a forum.Actor, in forum.ModerateInput) error {
	var user int
	var topicID int64
	var e error
	switch in.Target {
	case "topic":
		topicID = in.TargetID
		e = r.shared.pool.QueryRow(ctx, `SELECT user_id FROM forum_topics WHERE id=$1`, topicID).Scan(&user)
	case "post":
		e = r.shared.pool.QueryRow(ctx, `SELECT user_id,topic_id FROM forum_posts WHERE id=$1`, in.TargetID).Scan(&user, &topicID)
	default:
		return social.ErrForbidden
	}
	if e != nil {
		return forumNotFound(e)
	}
	tx, users, e := r.shared.begin(ctx, a, user)
	if e != nil {
		return e
	}
	defer rollbackSocialTx(tx)
	role := users[a.UserID].role
	if !(role == auth.RoleAdministrator || (role == auth.RoleModerator && (user == a.UserID || users[user].role == auth.RoleUser))) {
		return social.ErrForbidden
	}
	topic, e := loadForumTopic(ctx, tx, topicID, true)
	if e != nil {
		return e
	}
	var post *forum.Post
	if in.Target == "post" {
		post, e = loadForumPost(ctx, tx, in.TargetID, true)
		if e != nil {
			return e
		}
	}
	if e = checkSocialSession(ctx, tx, a); e != nil {
		return e
	}
	// Statements come from a fixed whitelist; no request text becomes SQL.
	var statement string
	if post != nil {
		switch in.Action {
		case "hide":
			if post.HiddenAt != nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_posts SET hidden_at=clock_timestamp(),hidden_by_user_id=$2 WHERE id=$1`
		case "restore":
			if post.HiddenAt == nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_posts SET hidden_at=NULL,hidden_by_user_id=NULL WHERE id=$1`
		default:
			return social.ErrForbidden
		}
	} else {
		switch in.Action {
		case "hide":
			if topic.HiddenAt != nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET hidden_at=clock_timestamp(),hidden_by_user_id=$2 WHERE id=$1`
		case "restore":
			if topic.HiddenAt == nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET hidden_at=NULL,hidden_by_user_id=NULL WHERE id=$1`
		case "close":
			if topic.ClosedAt != nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET closed_at=clock_timestamp() WHERE id=$1`
		case "open":
			if topic.ClosedAt == nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET closed_at=NULL WHERE id=$1`
		case "pin":
			if topic.PinnedAt != nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET pinned_at=clock_timestamp() WHERE id=$1`
		case "unpin":
			if topic.PinnedAt == nil {
				return forum.ErrConflict
			}
			statement = `UPDATE forum_topics SET pinned_at=NULL WHERE id=$1`
		default:
			return social.ErrForbidden
		}
	}
	args := []any{in.TargetID}
	if in.Action == "hide" {
		args = append(args, a.UserID)
	}
	if _, e = tx.Exec(ctx, statement, args...); e != nil {
		return e
	}
	var postID *int64
	if post != nil {
		postID = &in.TargetID
	}
	_, e = tx.Exec(ctx, `INSERT INTO moderation_actions(actor_user_id,user_id,scope,action,reason,forum_topic_id,forum_post_id) VALUES($1,$2,'forum',$3,$4,$5,$6)`, a.UserID, user, in.Action+"_"+in.Target, in.Reason, topicID, postID)
	if e != nil {
		return e
	}
	return tx.Commit(ctx)
}
