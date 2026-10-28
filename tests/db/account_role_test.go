package db_test

import (
	"gotest/internal/maintenance"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountRoleChangeIsExplicitAndAudited(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	if _, err := maintenance.SetAccountRole(ctx, s.pool, 1, "owner"); err == nil {
		t.Fatal("unknown role accepted")
	}
	change, err := maintenance.SetAccountRole(ctx, s.pool, 1, "administrator")
	if err != nil {
		t.Fatal(err)
	}
	if !change.Changed || change.Previous != "user" || change.Role != "administrator" {
		t.Fatalf("change=%+v", change)
	}
	change, err = maintenance.SetAccountRole(ctx, s.pool, 1, "administrator")
	if err != nil || change.Changed {
		t.Fatalf("repeated grant=%+v %v", change, err)
	}
	var role, count int
	if err := s.pool.QueryRow(ctx, `SELECT role_id FROM users WHERE id=1`).Scan(&role); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM moderation_actions WHERE user_id=1 AND action='role_change' AND actor_user_id IS NULL AND reason LIKE '%user%administrator%'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if role != 3 || count != 1 {
		t.Fatalf("role=%d audit=%d", role, count)
	}
	if _, err := maintenance.SetAccountRole(ctx, s.pool, 99999, "user"); err == nil {
		t.Fatal("missing user accepted")
	}
	mustExec(t, ctx, s, `ALTER TABLE moderation_actions ADD CONSTRAINT reject_new_audit CHECK(id=1)`)
	if _, err := maintenance.SetAccountRole(ctx, s.pool, 1, "user"); err == nil {
		t.Fatal("ignored audit failure")
	}
	if err := s.pool.QueryRow(ctx, `SELECT role_id FROM users WHERE id=1`).Scan(&role); err != nil {
		t.Fatal(err)
	}
	if role != 3 {
		t.Fatal("role changed without audit")
	}
}

func TestAccountRoleCommand(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	binary := filepath.Join(t.TempDir(), "accountrole.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/accountrole")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	dsn := s.pool.Config().ConnString()
	schema := s.pool.Config().ConnConfig.RuntimeParams["search_path"]
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("isolated test connection is invalid")
		}
		query := u.Query()
		query.Set("search_path", schema)
		u.RawQuery = query.Encode()
		dsn = u.String()
	} else {
		dsn += " search_path=" + schema
	}
	for _, tc := range []struct {
		args  []string
		valid bool
	}{{[]string{"-user-id", "1", "-role", "moderator"}, true}, {[]string{"-user-id", "1", "-role", "private-invalid-role"}, false}, {[]string{"-role", "administrator"}, false}} {
		cmd := exec.CommandContext(ctx, binary, tc.args...)
		cmd.Dir = t.TempDir()
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if !strings.EqualFold(key, "DATABASE_URL") {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		cmd.Env = append(cmd.Env, "DATABASE_URL="+dsn)
		output, err := cmd.CombinedOutput()
		if (err == nil) != tc.valid || strings.Contains(string(output), "private-invalid-role") || strings.Contains(string(output), "postgres://") {
			t.Fatalf("command outcome valid=%v err=%v", tc.valid, err)
		}
	}
	var role int
	if err := s.pool.QueryRow(ctx, `SELECT role_id FROM users WHERE id=1`).Scan(&role); err != nil {
		t.Fatal(err)
	}
	if role != 2 {
		t.Fatalf("role=%d", role)
	}
}
