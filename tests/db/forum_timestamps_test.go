package db_test

import (
	"testing"
	"time"
)

// Direct SQL must obey the same metadata contract as repository mutations.
func TestForumUpdatedAtTracksChangedValuesAndPreservesNoOps(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	mustExec(t, ctx, s, `INSERT INTO forum_topics(category_id,user_id,character_id,author_name,title,body) VALUES(2,1,1,'author','original topic','original body');
		INSERT INTO forum_posts(topic_id,user_id,character_id,author_name,body) VALUES(1,1,1,'author','original reply')`)
	for _, tc := range []struct{ table, field, changed string }{
		{"forum_categories", "description", "changed category description"},
		{"forum_topics", "body", "changed topic body"},
		{"forum_posts", "body", "changed reply body"},
	} {
		t.Run(tc.table, func(t *testing.T) {
			var created, original, changed, unchanged time.Time
			if e := s.pool.QueryRow(ctx, "SELECT created_at,updated_at FROM "+tc.table+" WHERE id=1").Scan(&created, &original); e != nil {
				t.Fatal(e)
			}
			if e := s.pool.QueryRow(ctx, "UPDATE "+tc.table+" SET "+tc.field+"=$1 WHERE id=1 RETURNING updated_at", tc.changed).Scan(&changed); e != nil {
				t.Fatal(e)
			}
			if !changed.After(original) {
				t.Fatalf("changed row did not advance updated_at: original=%v changed=%v", original, changed)
			}
			var afterCreated time.Time
			if e := s.pool.QueryRow(ctx, "UPDATE "+tc.table+" SET "+tc.field+"=$1 WHERE id=1 RETURNING created_at,updated_at", tc.changed).Scan(&afterCreated, &unchanged); e != nil {
				t.Fatal(e)
			}
			if !unchanged.Equal(changed) || !afterCreated.Equal(created) {
				t.Fatalf("no-op changed timestamps: created=%v want=%v updated=%v want=%v", afterCreated, created, unchanged, changed)
			}
		})
	}
}
