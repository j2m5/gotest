package db_test

import "testing"

const forumMigration = "20261029000100_forum"

func TestForumSchemaAndReversibleMigrationPreserveLegacyJournal(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM forum_categories`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("forum categories=%d %v", count, err)
	}
	mustExec(t, ctx, s, `INSERT INTO moderation_actions(actor_user_id,user_id,scope,action,reason) VALUES(2,1,'forum','ban','old forum ban'),(2,1,'chat','role_change','old role event');
	 INSERT INTO chat_messages(user_id,character_id,author_name,author_role,body) VALUES(1,1,'original chat author','user','original chat body');
	 INSERT INTO moderation_actions(actor_user_id,user_id,scope,action,reason,message_id) VALUES(2,1,'chat','delete_message','old chat event',1);
 INSERT INTO forum_topics(category_id,user_id,character_id,author_name,title,body) VALUES(2,1,1,'name','title','body');
 INSERT INTO forum_posts(topic_id,user_id,character_id,author_name,body) VALUES(1,1,1,'name','body');
 INSERT INTO moderation_actions(actor_user_id,user_id,scope,action,reason,forum_topic_id,forum_post_id) VALUES(2,1,'forum','hide_post','new event',1,1)`)
	for _, sql := range []string{
		`UPDATE forum_topics SET title='ab' WHERE id=1`,
		`UPDATE forum_topics SET body='' WHERE id=1`,
		`UPDATE forum_topics SET version=0 WHERE id=1`,
		`UPDATE forum_posts SET body=repeat('я',10001) WHERE id=1`,
		`UPDATE forum_posts SET version=0 WHERE id=1`,
		`UPDATE forum_topics SET character_id=3 WHERE id=1`,
		`UPDATE forum_posts SET character_id=3 WHERE id=1`,
	} {
		if _, e := s.pool.Exec(ctx, sql); e == nil {
			t.Fatalf("accepted invalid forum data: %s", sql)
		}
	}
	executeMigration(t, ctx, s, forumMigration+".down.sql")
	if e := s.pool.QueryRow(ctx, `SELECT count(*) FROM moderation_actions`).Scan(&count); e != nil || count != 3 {
		t.Fatalf("legacy audit=%d %v", count, e)
	}
	var messageID int64
	if e := s.pool.QueryRow(ctx, `SELECT message_id FROM moderation_actions WHERE action='delete_message'`).Scan(&messageID); e != nil || messageID != 1 {
		t.Fatalf("legacy message reference=%d %v", messageID, e)
	}
	if e := s.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); e != nil || count != 2 {
		t.Fatalf("users=%d %v", count, e)
	}
	executeMigration(t, ctx, s, forumMigration+".up.sql")
	if e := s.pool.QueryRow(ctx, `SELECT count(*) FROM forum_categories`).Scan(&count); e != nil || count != 4 {
		t.Fatalf("recreated categories=%d %v", count, e)
	}
}
