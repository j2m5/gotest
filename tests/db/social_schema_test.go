package db_test

import "testing"

func TestSocialSchemaRolesAndBanBounds(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var roles int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM account_roles WHERE (id,name) IN ((1,'user'),(2,'moderator'),(3,'administrator'))`).Scan(&roles); err != nil {
		t.Fatal(err)
	}
	if roles != 3 {
		t.Fatalf("roles=%d", roles)
	}
	for _, expiry := range []string{"NULL", "clock_timestamp()+interval '30 minutes'", "clock_timestamp()+interval '169 hours'"} {
		if _, err := s.pool.Exec(ctx, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,expires_at) VALUES(1,'chat','reason',2,`+expiry+`)`); err == nil {
			t.Fatalf("accepted invalid expiry %s", expiry)
		}
	}
	for _, sql := range []string{
		`INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','x',2)`,
		`INSERT INTO chat_messages(user_id,character_id,author_name,author_role,body) VALUES(1,1,'name','user','')`,
		`INSERT INTO chat_messages(user_id,character_id,author_name,author_role,body) VALUES(1,1,'name','user',repeat('я',1001))`,
		`UPDATE users SET role_id=99 WHERE id=1`,
	} {
		if _, err := s.pool.Exec(ctx, sql); err == nil {
			t.Fatalf("constraint accepted %s", sql)
		}
	}
}

func TestSocialMigrationsDownAndUp(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, name := range []string{"20261028000200_chat.down.sql", "20261028000100_social_moderation.down.sql", "20261028000100_social_moderation.up.sql", "20261028000200_chat.up.sql"} {
		executeMigration(t, ctx, s, name)
	}
}
