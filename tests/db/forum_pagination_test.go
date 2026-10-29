package db_test

import (
	"gotest/internal/forum"
	"testing"
)

func TestForumPinnedKeysetAndPostPages(t *testing.T) {
	s, ctx, r, _, user, mod := forumActors(t)
	mustExec(t, ctx, s, `INSERT INTO forum_topics(category_id,user_id,character_id,author_name,title,body,pinned_at) SELECT 2,1,1,'snapshot','topic '||g,'opening '||g,CASE WHEN g<=2 THEN clock_timestamp() ELSE NULL END FROM generate_series(1,34) g;
 INSERT INTO forum_topics(category_id,user_id,character_id,author_name,title,body) VALUES(3,1,1,'snapshot','other category','body');
 INSERT INTO forum_posts(topic_id,user_id,character_id,author_name,body) SELECT 1,1,1,'snapshot','post '||g FROM generate_series(1,33) g`)
	list, e := r.Topics(ctx, user, forum.TopicsInput{CategoryID: 2})
	if e != nil || len(list.Topics) != 30 || list.NextCursor == nil || *list.NextCursor != "n:7" {
		t.Fatalf("first topics=%+v %v", list, e)
	}
	if list.Topics[0].ID != 2 || list.Topics[1].ID != 1 || list.Topics[2].ID != 34 || list.Topics[29].ID != 7 {
		t.Fatalf("topic order=%+v", list.Topics)
	}
	next, e := r.Topics(ctx, user, forum.TopicsInput{CategoryID: 2, Cursor: *list.NextCursor})
	if e != nil || len(next.Topics) != 4 || next.Topics[0].ID != 6 || next.Topics[3].ID != 3 || next.NextCursor != nil {
		t.Fatalf("next topics=%+v %v", next, e)
	}
	pinned, e := r.Topics(ctx, user, forum.TopicsInput{CategoryID: 2, Cursor: "p:2"})
	if e != nil || len(pinned.Topics) != 30 || pinned.Topics[0].ID != 1 || pinned.Topics[1].ID != 34 {
		t.Fatalf("pin transition=%+v %v", pinned, e)
	}
	moderateForum(t, ctx, r, mod, "post", 2, "hide")
	page, e := r.Topic(ctx, user, 1, 0)
	if e != nil || len(page.Posts) != 30 || page.NextAfterID == nil || *page.NextAfterID != 30 || page.Posts[1].Body != nil || page.Topic.Body == nil || page.Topic.ReplyCount != 33 {
		t.Fatalf("first posts=%+v %v", page, e)
	}
	page, e = r.Topic(ctx, user, 1, *page.NextAfterID)
	if e != nil || len(page.Posts) != 3 || page.Posts[0].ID != 31 || page.NextAfterID != nil || page.Topic.ID != 1 {
		t.Fatalf("next posts=%+v %v", page, e)
	}
}
