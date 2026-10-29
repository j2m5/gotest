package db_test

import (
	"errors"
	"gotest/internal/auth"
	"gotest/internal/forum"
	"gotest/internal/social"
	"testing"
)

func TestForumQueuedWriteRechecksCurrentAccountPolicy(t *testing.T) {
	for _, change := range []string{"forum ban", "expired forum ban", "account ban", "session", "character", "no character"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, r, _, user, _ := forumActors(t)
			topic := forumTopic(t, ctx, r, user)
			ageForum(t, ctx, s)
			tx, pid := lockResetUser(t, ctx, s)
			defer tx.Rollback(ctx)
			done := make(chan error, 1)
			go func() {
				_, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "pending reply"})
				done <- e
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			// A single timestamp respects the ban's exact whole-hour CHECK.
			sql := `WITH stamp AS(SELECT clock_timestamp() AS at) INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,created_at,expires_at) SELECT 1,'forum','new ban',2,at,at+interval '1 hour' FROM stamp`
			want := error(forum.ErrBanned)
			switch change {
			case "expired forum ban":
				sql = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,created_at,expires_at) VALUES(1,'forum','expired ban',2,now()-interval '2 hours',now()-interval '1 hour')`
				want = nil
			case "account ban":
				sql = `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','new ban',2)`
				want = auth.ErrUnauthenticated
			case "session":
				sql = `DELETE FROM sessions WHERE user_id=1`
				want = auth.ErrUnauthenticated
			case "character":
				sql = `UPDATE users SET character_id=2 WHERE id=1`
				want = social.ErrCharacterChanged
			case "no character":
				sql = `UPDATE users SET character_id=NULL WHERE id=1`
				want = social.ErrNoCharacter
			}
			if _, e := tx.Exec(ctx, sql); e != nil {
				t.Fatal(e)
			}
			if e := tx.Commit(ctx); e != nil {
				t.Fatal(e)
			}
			if e := <-done; !errors.Is(e, want) {
				t.Fatalf("queued write=%v want=%v", e, want)
			}
		})
	}
}

func TestForumQueuedModeratorRechecksRoleAndSession(t *testing.T) {
	for _, change := range []string{"demoted actor", "promoted target", "session"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, r, _, user, mod := forumActors(t)
			topic := forumTopic(t, ctx, r, user)
			tx, pid := lockResetUser(t, ctx, s)
			defer tx.Rollback(ctx)
			done := make(chan error, 1)
			go func() {
				done <- r.Moderate(ctx, mod, forum.ModerateInput{Target: "topic", TargetID: topic.ID, Action: "close", Reason: "pending moderation"})
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			sql := `UPDATE users SET role_id=1 WHERE id=2`
			want := error(social.ErrForbidden)
			if change == "promoted target" {
				sql = `UPDATE users SET role_id=2 WHERE id=1`
			}
			if change == "session" {
				sql = `DELETE FROM sessions WHERE user_id=2`
				want = auth.ErrUnauthenticated
			}
			if _, e := tx.Exec(ctx, sql); e != nil {
				t.Fatal(e)
			}
			if e := tx.Commit(ctx); e != nil {
				t.Fatal(e)
			}
			if e := <-done; !errors.Is(e, want) {
				t.Fatalf("queued moderation=%v want=%v", e, want)
			}
		})
	}
}

func TestForumReplyWaitsForTopicCloseAndHide(t *testing.T) {
	for _, action := range []string{"close", "hide"} {
		t.Run(action, func(t *testing.T) {
			s, ctx, r, _, user, _ := forumActors(t)
			topic := forumTopic(t, ctx, r, user)
			ageForum(t, ctx, s)
			tx, e := s.pool.Begin(ctx)
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback(ctx)
			var pid int
			if e = tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM forum_topics WHERE id=$1 FOR UPDATE`, topic.ID).Scan(&pid); e != nil {
				t.Fatal(e)
			}
			done := make(chan error, 1)
			go func() {
				_, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "racing reply"})
				done <- e
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			sql := `UPDATE forum_topics SET closed_at=clock_timestamp() WHERE id=$1`
			want := error(forum.ErrClosed)
			if action == "hide" {
				sql = `UPDATE forum_topics SET hidden_at=clock_timestamp(),hidden_by_user_id=2 WHERE id=$1`
				want = forum.ErrNotFound
			}
			if _, e = tx.Exec(ctx, sql, topic.ID); e != nil {
				t.Fatal(e)
			}
			if e = tx.Commit(ctx); e != nil {
				t.Fatal(e)
			}
			if e = <-done; !errors.Is(e, want) {
				t.Fatalf("racing reply=%v want=%v", e, want)
			}
			var count int
			if e = s.pool.QueryRow(ctx, `SELECT count(*) FROM forum_posts`).Scan(&count); e != nil || count != 0 {
				t.Fatalf("post bypass=%d %v", count, e)
			}
		})
	}
}

func TestForumConcurrentWritesShareAccountThrottle(t *testing.T) {
	s, ctx, r, _, user, _ := forumActors(t)
	topic := forumTopic(t, ctx, r, user)
	ageForum(t, ctx, s)
	tx, pid := lockResetUser(t, ctx, s)
	defer tx.Rollback(ctx)
	done := make(chan error, 2)
	go func() {
		_, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "reply"})
		done <- e
	}()
	go func() {
		_, e := r.Create(ctx, user, forum.CreateInput{CategoryID: 2, CharacterID: 1, Title: "New topic", Body: "body"})
		done <- e
	}()
	awaitResetUserLock(t, ctx, s, pid, done)
	if e := tx.Commit(ctx); e != nil {
		t.Fatal(e)
	}
	first, second := <-done, <-done
	if !((first == nil && errors.Is(second, forum.ErrTooFast)) || (second == nil && errors.Is(first, forum.ErrTooFast))) {
		t.Fatalf("concurrent writes=%v, %v", first, second)
	}
}

func TestForumSessionExpiryWhileWaitingForTopicLock(t *testing.T) {
	for _, operation := range []string{"reply", "edit topic", "edit post", "moderate"} {
		t.Run(operation, func(t *testing.T) {
			s, ctx, r, _, user, mod := forumActors(t)
			topic := forumTopic(t, ctx, r, user)
			ageForum(t, ctx, s)
			post, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "original reply"})
			if e != nil {
				t.Fatal(e)
			}
			ageForum(t, ctx, s)
			tx, e := s.pool.Begin(ctx)
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback(ctx)
			var pid int
			if e = tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM forum_topics WHERE id=$1 FOR UPDATE`, topic.ID).Scan(&pid); e != nil {
				t.Fatal(e)
			}
			done := make(chan error, 1)
			go func() {
				var err error
				switch operation {
				case "reply":
					_, err = r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "pending"})
				case "edit topic":
					_, err = r.EditTopic(ctx, user, forum.EditTopicInput{TopicID: topic.ID, CharacterID: 1, Title: "Pending edit", Body: "pending", Version: 1})
				case "edit post":
					_, err = r.EditPost(ctx, user, forum.EditPostInput{PostID: post.ID, CharacterID: 1, Body: "pending", Version: 1})
				case "moderate":
					err = r.Moderate(ctx, mod, forum.ModerateInput{Target: "post", TargetID: post.ID, Action: "hide", Reason: "pending moderation"})
				}
				done <- err
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			actor := user
			if operation == "moderate" {
				actor = mod
			}
			// Moving the expiry into the past is a deterministic substitute for
			// waiting for a real token deadline while the content row is locked.
			if _, e = tx.Exec(ctx, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=$1`, actor.UserID); e != nil {
				t.Fatal(e)
			}
			if e = tx.Commit(ctx); e != nil {
				t.Fatal(e)
			}
			if e = <-done; !errors.Is(e, auth.ErrUnauthenticated) {
				t.Fatalf("expired waiting session=%v", e)
			}
		})
	}
}
