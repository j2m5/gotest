package db_test

import (
	"context"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/forum"
	"gotest/internal/social"
	"testing"
)

func forumActors(t *testing.T) (*testStorage, context.Context, *db.ForumRepository, *db.SocialRepository, forum.Actor, forum.Actor) {
	t.Helper()
	s, ctx, socialRepo, user, mod := socialActors(t)
	mustExec(t, ctx, s, `UPDATE users SET character_id=3 WHERE id=2`)
	return s, ctx, db.NewForumRepository(s.pool), socialRepo, user, mod
}
func forumTopic(t *testing.T, ctx context.Context, r *db.ForumRepository, a forum.Actor) *forum.Topic {
	t.Helper()
	v, e := r.Create(ctx, a, forum.CreateInput{CategoryID: 2, CharacterID: 1, Title: "A topic", Body: "<script>literal</script>"})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func ageForum(t *testing.T, ctx context.Context, s *testStorage) {
	t.Helper()
	mustExec(t, ctx, s, `UPDATE forum_topics SET created_at=clock_timestamp()-interval '10 seconds';UPDATE forum_posts SET created_at=clock_timestamp()-interval '10 seconds'`)
}
func moderateForum(t *testing.T, ctx context.Context, r *db.ForumRepository, a forum.Actor, target string, id int64, action string) {
	t.Helper()
	if e := r.Moderate(ctx, a, forum.ModerateInput{Target: target, TargetID: id, Action: action, Reason: "test reason"}); e != nil {
		t.Fatal(e)
	}
}

func TestForumScopeBansAndCharacterAuthorship(t *testing.T) {
	s, ctx, r, socialRepo, user, mod := forumActors(t)
	categories, e := r.Categories(ctx, user)
	if e != nil || len(categories.Categories) != 4 || categories.Categories[0].CreateRole != "administrator" {
		t.Fatalf("categories=%+v %v", categories, e)
	}
	if _, e = r.Create(ctx, user, forum.CreateInput{CategoryID: 1, CharacterID: 1, Title: "News topic", Body: "text"}); !errors.Is(e, social.ErrForbidden) {
		t.Fatalf("news permission=%v", e)
	}
	topic := forumTopic(t, ctx, r, user)
	if topic.CharacterID != 1 || topic.AuthorName != "First character" || topic.Body == nil || *topic.Body != "<script>literal</script>" || topic.Version != 1 {
		t.Fatalf("topic=%+v", topic)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1;UPDATE characters SET name='Renamed first' WHERE id=1`)
	if _, e = r.EditTopic(ctx, user, forum.EditTopicInput{TopicID: topic.ID, CharacterID: 1, Title: "updated", Body: "text", Version: 1}); !errors.Is(e, social.ErrCharacterChanged) {
		t.Fatalf("stale character=%v", e)
	}
	edited, e := r.EditTopic(ctx, user, forum.EditTopicInput{TopicID: topic.ID, CharacterID: 2, Title: "updated", Body: "edited", Version: 1})
	if e != nil || edited.CharacterID != 1 || edited.AuthorName != "First character" || edited.Version != 2 || edited.EditedAt == nil || !edited.CreatedAt.Equal(topic.CreatedAt) || !edited.UpdatedAt.After(topic.UpdatedAt) {
		t.Fatalf("edit=%+v %v", edited, e)
	}
	if _, e = r.EditTopic(ctx, user, forum.EditTopicInput{TopicID: topic.ID, CharacterID: 2, Title: "stale", Body: "lost", Version: 1}); !errors.Is(e, forum.ErrConflict) {
		t.Fatalf("version=%v", e)
	}
	h := 1
	ban, e := socialRepo.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "forum", DurationHours: &h, Reason: "forum reason"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 2, Body: "blocked"}); !errors.Is(e, forum.ErrBanned) {
		t.Fatalf("forum ban reply=%v", e)
	}
	if _, e = r.EditTopic(ctx, user, forum.EditTopicInput{TopicID: topic.ID, CharacterID: 2, Title: "blocked", Body: "blocked", Version: 2}); !errors.Is(e, forum.ErrBanned) {
		t.Fatalf("forum ban edit=%v", e)
	}
	if _, e = socialRepo.Send(ctx, user, social.SendInput{CharacterID: 2, Body: "chat allowed"}); e != nil {
		t.Fatalf("forum ban blocked chat=%v", e)
	}
	page, e := r.Topic(ctx, user, topic.ID, 0)
	if e != nil || page.ForumBan == nil || page.ForumBan.ID != ban.ID || page.Posts == nil {
		t.Fatalf("banned reading=%+v %v", page, e)
	}
	if e = socialRepo.Revoke(ctx, mod, social.RevokeInput{BanID: ban.ID, Reason: "lift ban"}); e != nil {
		t.Fatal(e)
	}
	if _, e = socialRepo.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "chat", DurationHours: &h, Reason: "chat reason"}); e != nil {
		t.Fatal(e)
	}
	ageForum(t, ctx, s)
	post, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 2, Body: "forum allowed"})
	if e != nil || post.CharacterID != 2 || post.AuthorName != "Second character" {
		t.Fatalf("chat ban blocked forum=%+v %v", post, e)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=NULL WHERE id=1`)
	if _, e = r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 2, Body: "none"}); !errors.Is(e, social.ErrNoCharacter) {
		t.Fatalf("no character=%v", e)
	}
	if _, e = r.Categories(ctx, user); e != nil {
		t.Fatalf("read without character=%v", e)
	}
	mustExec(t, ctx, s, `UPDATE sessions SET expires_at=clock_timestamp()-interval '1 second' WHERE user_id=1`)
	if _, e = r.Categories(ctx, user); !errors.Is(e, auth.ErrUnauthenticated) {
		t.Fatalf("expired session=%v", e)
	}
}

func TestForumHiddenDirectReadsRestoreAndVersionedPosts(t *testing.T) {
	s, ctx, r, socialRepo, user, mod := forumActors(t)
	topic := forumTopic(t, ctx, r, user)
	ageForum(t, ctx, s)
	post, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "secret reply"})
	if e != nil {
		t.Fatal(e)
	}
	edited, e := r.EditPost(ctx, user, forum.EditPostInput{PostID: post.ID, CharacterID: 1, Body: "edited secret", Version: 1})
	if e != nil || edited.Version != 2 || edited.EditedAt == nil || !edited.CreatedAt.Equal(post.CreatedAt) || !edited.UpdatedAt.After(post.UpdatedAt) {
		t.Fatalf("post edit=%+v %v", edited, e)
	}
	if _, e = r.EditPost(ctx, user, forum.EditPostInput{PostID: post.ID, CharacterID: 1, Body: "stale", Version: 1}); !errors.Is(e, forum.ErrConflict) {
		t.Fatalf("post version=%v", e)
	}
	moderateForum(t, ctx, r, mod, "post", post.ID, "hide")
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "hide")
	if _, e = r.Topic(ctx, user, topic.ID, 0); !errors.Is(e, forum.ErrNotFound) {
		t.Fatalf("hidden direct read=%v", e)
	}
	if _, e = r.EditPost(ctx, user, forum.EditPostInput{PostID: post.ID, CharacterID: 1, Body: "bypass hidden parent", Version: 2}); !errors.Is(e, forum.ErrNotFound) {
		t.Fatalf("hidden parent edit=%v", e)
	}
	page, e := r.Topic(ctx, mod, topic.ID, 0)
	if e != nil || page.Topic.Body == nil || page.Posts[0].Body == nil || *page.Posts[0].Body != "edited secret" {
		t.Fatalf("staff hidden=%+v %v", page, e)
	}
	list, e := r.Topics(ctx, mod, forum.TopicsInput{CategoryID: 2})
	if e != nil || len(list.Topics) != 0 || list.Topics == nil {
		t.Fatalf("default hidden list=%+v %v", list, e)
	}
	list, e = r.Topics(ctx, mod, forum.TopicsInput{CategoryID: 2, IncludeHidden: true})
	if e != nil || len(list.Topics) != 1 {
		t.Fatalf("staff list=%+v %v", list, e)
	}
	if _, e = r.Topics(ctx, user, forum.TopicsInput{CategoryID: 2, IncludeHidden: true}); !errors.Is(e, social.ErrForbidden) {
		t.Fatalf("hidden flag=%v", e)
	}
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "restore")
	page, e = r.Topic(ctx, user, topic.ID, 0)
	if e != nil || len(page.Posts) != 1 || page.Posts[0].Body != nil || page.Posts[0].HiddenAt == nil || page.Topic.ReplyCount != 1 {
		t.Fatalf("restored placeholders=%+v %v", page, e)
	}
	moderateForum(t, ctx, r, mod, "post", post.ID, "restore")
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "close")
	if _, e = r.EditPost(ctx, user, forum.EditPostInput{PostID: post.ID, CharacterID: 1, Body: "closed bypass", Version: 2}); !errors.Is(e, forum.ErrClosed) {
		t.Fatalf("closed edit=%v", e)
	}
	if _, e = r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "closed"}); !errors.Is(e, forum.ErrClosed) {
		t.Fatalf("closed reply=%v", e)
	}
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "open")
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "pin")
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "unpin")
	actions, e := socialRepo.Actions(ctx, mod, 0)
	if e != nil || len(actions.Actions) != 8 {
		t.Fatalf("audit=%+v %v", actions, e)
	}
	for _, a := range actions.Actions {
		if a.Scope != "forum" || a.ForumTopicID == nil || *a.ForumTopicID != topic.ID || a.MessageID != nil {
			t.Fatalf("forum audit=%+v", a)
		}
		if a.Action == "hide_post" && (a.ForumPostID == nil || *a.ForumPostID != post.ID) {
			t.Fatalf("post audit=%+v", a)
		}
	}
}

func TestForumHierarchyCurrentRolesAndAtomicAudit(t *testing.T) {
	s, ctx, r, _, user, mod := forumActors(t)
	topic := forumTopic(t, ctx, r, user)
	mustExec(t, ctx, s, `UPDATE users SET role_id=2 WHERE id=1`)
	if e := r.Moderate(ctx, mod, forum.ModerateInput{Target: "topic", TargetID: topic.ID, Action: "hide", Reason: "peer forbidden"}); !errors.Is(e, social.ErrForbidden) {
		t.Fatalf("peer moderation=%v", e)
	}
	moderateForum(t, ctx, r, user, "topic", topic.ID, "pin")
	page, e := r.Topic(ctx, mod, topic.ID, 0)
	if e != nil || page.Topic.AuthorRole != "moderator" {
		t.Fatalf("current author role=%+v %v", page, e)
	}
	mustExec(t, ctx, s, `UPDATE users SET role_id=3 WHERE id=2;ALTER TABLE moderation_actions ADD CONSTRAINT reject_forum_audit CHECK(action NOT IN ('hide_topic'))`)
	if e = r.Moderate(ctx, mod, forum.ModerateInput{Target: "topic", TargetID: topic.ID, Action: "hide", Reason: "must rollback"}); e == nil {
		t.Fatal("ignored audit failure")
	}
	page, e = r.Topic(ctx, user, topic.ID, 0)
	if e != nil || page.Topic.HiddenAt != nil {
		t.Fatalf("partial hide=%+v %v", page, e)
	}
	mustExec(t, ctx, s, `ALTER TABLE moderation_actions DROP CONSTRAINT reject_forum_audit`)
	moderateForum(t, ctx, r, mod, "topic", topic.ID, "hide")
}

func TestForumCombinedAccountThrottleCountsHiddenWrites(t *testing.T) {
	s, ctx, r, _, user, mod := forumActors(t)
	topic := forumTopic(t, ctx, r, user)
	if _, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "too fast"}); !errors.Is(e, forum.ErrTooFast) {
		t.Fatalf("topic to reply throttle=%v", e)
	}
	ageForum(t, ctx, s)
	post, e := r.Reply(ctx, user, forum.ReplyInput{TopicID: topic.ID, CharacterID: 1, Body: "reply"})
	if e != nil {
		t.Fatal(e)
	}
	moderateForum(t, ctx, r, mod, "post", post.ID, "hide")
	if _, e = r.Create(ctx, user, forum.CreateInput{CategoryID: 2, CharacterID: 1, Title: "Another topic", Body: "too fast"}); !errors.Is(e, forum.ErrTooFast) {
		t.Fatalf("hidden reply to topic throttle=%v", e)
	}
}
