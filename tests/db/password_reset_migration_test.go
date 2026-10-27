package db_test

import (
	"testing"

	"gotest/internal/auth"
)

func TestPasswordResetMigrationUpDownUp(t *testing.T) {
	const migration = "20261027000100_password_resets"
	s, ctx := characterTestStorageBefore(t, migration)
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1`)
	for _, direction := range []string{"up", "down", "up"} {
		executeMigration(t, ctx, s, migration+"."+direction+".sql")
		s.pool.Reset()
		var intact bool
		if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users)=2 AND (SELECT count(*) FROM characters)=3
			AND EXISTS(SELECT 1 FROM users WHERE id=1 AND password='unused' AND character_id=1 AND email_verified_at IS NOT NULL)`).Scan(&intact); err != nil || !intact {
			t.Fatalf("%s changed existing accounts: %v %v", direction, intact, err)
		}
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT to_regclass(current_schema()||'.password_resets') IS NOT NULL`).Scan(&exists); err != nil || exists != (direction == "up") {
			t.Fatalf("%s table exists=%v err=%v", direction, exists, err)
		}
		if direction == "down" {
			continue
		}
		var count int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets`).Scan(&count); err != nil || count != 0 {
			t.Fatalf("up restored old links: %d %v", count, err)
		}
		createResetLink(t, ctx, s, "first@example.test", "migration-link")
		var schemaOK bool
		if err := s.pool.QueryRow(ctx, `SELECT
			(SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='password_resets' AND column_name IN ('created_at','expires_at') AND data_type='timestamp with time zone')=2
			AND EXISTS(SELECT 1 FROM pg_indexes WHERE schemaname=current_schema() AND tablename='password_resets' AND indexdef LIKE '%(expires_at, id)%')
			AND EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='password_resets'::regclass AND contype='p')`).Scan(&schemaOK); err != nil || !schemaOK {
			t.Fatalf("reset schema: %v %v", schemaOK, err)
		}
		for _, sql := range []string{
			`INSERT INTO password_resets(user_id,token,expires_at) VALUES(1,'` + auth.TokenDigest("duplicate-user") + `',clock_timestamp()+interval '1 hour')`,
			`INSERT INTO password_resets(user_id,token,expires_at) VALUES(2,'` + auth.TokenDigest("migration-link") + `',clock_timestamp()+interval '1 hour')`,
			`INSERT INTO password_resets(user_id,token,expires_at) VALUES(2,'raw-token',clock_timestamp()+interval '1 hour')`,
			`INSERT INTO password_resets(user_id,token,expires_at) VALUES(99999,'` + auth.TokenDigest("missing-user") + `',clock_timestamp()+interval '1 hour')`,
		} {
			if _, err := s.pool.Exec(ctx, sql); err == nil {
				t.Fatal("reset constraint accepted invalid row")
			}
		}
	}
	// A separate account can be deleted without deleting its character selection first.
	createResetLink(t, ctx, s, "second@example.test", "cascade-link")
	mustExec(t, ctx, s, `DELETE FROM users WHERE id=2`)
	var remaining int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets WHERE user_id=2`).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("account cascade left links: %d %v", remaining, err)
	}
}
