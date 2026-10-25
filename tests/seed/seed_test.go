package seed_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const codeMigration = "20261025000200_seed_codes"

var seedBinary string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		return m.Run()
	}
	dir, err := os.MkdirTemp("", "gotest-seed-binary-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	seedBinary = filepath.Join(dir, "seed.exe")
	cmd := exec.Command("go", "build", "-o", seedBinary, "./cmd/seed")
	cmd.Dir = filepath.Join("..", "..")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build seed: %v\n%s", err, output)
		return 1
	}
	return m.Run()
}

type seedDB struct {
	pool *pgxpool.Pool
	ctx  context.Context
	dsn  string
	dir  string
}

// Every test owns a schema in an explicitly selected TEST_DATABASE_URL.
// Never fall back to the application's database or load its .env.
func newSeedDB(t *testing.T, legacy bool) *seedDB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is unset; seed integration tests require PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	var nonce [12]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	schema := "seed_test_" + hex.EncodeToString(nonce[:])
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		dsn = u.String()
	} else {
		dsn += " search_path=" + schema
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	s := &seedDB{pool: pool, ctx: ctx, dsn: dsn, dir: t.TempDir()}
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("find migrations: %v, files=%d", err, len(files))
	}
	for _, path := range files {
		if legacy && filepath.Base(path) >= codeMigration+".up.sql" {
			continue
		}
		s.execFile(t, path)
	}
	return s
}

func (s *seedDB) execFile(t *testing.T, path string) {
	t.Helper()
	s.exec(t, readFile(t, path))
}

func (s *seedDB) exec(t *testing.T, sql string) {
	t.Helper()
	if _, err := s.pool.Exec(s.ctx, sql); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeFile(t *testing.T, path, sql string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(sql), 0600); err != nil {
		t.Fatal(err)
	}
}

func (s *seedDB) copySeeds(t *testing.T, name string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "seeds", "*.sql"))
	if err != nil || len(files) != 6 {
		t.Fatalf("find seeds: %v, files=%d", err, len(files))
	}
	for _, path := range files {
		writeFile(t, filepath.Join(s.dir, name, filepath.Base(path)), readFile(t, path))
	}
}

func (s *seedDB) run(extraEnv []string, args ...string) (string, error) {
	cmd := exec.CommandContext(s.ctx, seedBinary, args...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), "DATABASE_URL="+s.dsn, "APP_ENV=local", "APP_PORT=8081", "PUBLIC_URL=http://localhost:8081", "SMTP_TIMEOUT=")
	cmd.Env = append(cmd.Env, extraEnv...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *seedDB) assertCounts(t *testing.T, want [6]int) {
	t.Helper()
	var got [6]int
	err := s.pool.QueryRow(s.ctx, `SELECT
		(SELECT count(*) FROM factions), (SELECT count(*) FROM classes),
		(SELECT count(*) FROM abilities), (SELECT count(*) FROM class_roles),
		(SELECT count(*) FROM class_abilities), (SELECT count(*) FROM levels)`).Scan(
		&got[0], &got[1], &got[2], &got[3], &got[4], &got[5])
	if err != nil || got != want {
		t.Fatalf("catalog counts = %v, want %v; err=%v", got, want, err)
	}
}

func (s *seedDB) snapshot(t *testing.T) string {
	t.Helper()
	var result string
	err := s.pool.QueryRow(s.ctx, `SELECT jsonb_build_array(
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM factions t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM classes t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM abilities t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY class_id,role) FROM class_roles t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY class_id,ability_id) FROM class_abilities t),
		(SELECT jsonb_agg(to_jsonb(t) ORDER BY level) FROM levels t))::text`).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSeedRepeatablePreservesExistingValues(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "seeds")
	if output, err := s.run(nil); err != nil {
		t.Fatalf("first seed: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{2, 8, 8, 20, 8, 50})
	s.exec(t, "UPDATE abilities SET damage=123, name='Edited ability'; UPDATE levels SET experience_required=999 WHERE level=2")
	before := s.snapshot(t)
	if output, err := s.run(nil); err != nil {
		t.Fatalf("second seed must succeed: %v\n%s", err, output)
	}
	if after := s.snapshot(t); after != before {
		t.Fatal("repeated seed changed existing rows, IDs, timestamps or game balance")
	}
}

func TestSeedDoesNotAssumeIDs(t *testing.T) {
	s := newSeedDB(t, false)
	s.exec(t, `SELECT setval('factions_id_seq',100); SELECT setval('classes_id_seq',200); SELECT setval('abilities_id_seq',300)`)
	s.copySeeds(t, "seeds")
	if output, err := s.run(nil); err != nil {
		t.Fatalf("seed with advanced sequences: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{2, 8, 8, 20, 8, 50})
	var faction, icon string
	err := s.pool.QueryRow(s.ctx, `SELECT f.name,a.icon FROM classes c
		JOIN factions f ON f.id=c.faction_id JOIN class_abilities ca ON ca.class_id=c.id
		JOIN abilities a ON a.id=ca.ability_id WHERE c.name='Jedi Consular'`).Scan(&faction, &icon)
	if err != nil || faction != "Galactic Republic" || icon != "images/abilities/1720182283_1756577701.png" {
		t.Fatalf("incorrect references: faction=%q icon=%q err=%v", faction, icon, err)
	}
}

func TestSeedRollsBackWholeBatch(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "seeds")
	writeFile(t, filepath.Join(s.dir, "seeds", "99_failure.sql"), "SELECT missing_seed_function()")
	output, err := s.run(nil)
	if err == nil || !strings.Contains(output, "99_failure.sql") {
		t.Fatalf("expected failure identifying its file: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{})
}

func TestSeedRejectsEmptyDirectory(t *testing.T) {
	s := newSeedDB(t, false)
	if err := os.Mkdir(filepath.Join(s.dir, "seeds"), 0700); err != nil {
		t.Fatal(err)
	}
	if output, err := s.run(nil); err == nil {
		t.Fatalf("empty directory reported success: %s", output)
	}
}

func TestSeedIndependentOfHTTPAndMail(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "seeds")
	if output, err := s.run([]string{"APP_PORT=invalid", "PUBLIC_URL=invalid", "SMTP_TIMEOUT=invalid", "MAIL_DRIVER=invalid"}); err != nil {
		t.Fatalf("unrelated settings must not block seed: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{2, 8, 8, 20, 8, 50})
}

func TestSeedCustomDirectory(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "catalog data")
	if output, err := s.run(nil, "-dir", filepath.Join(s.dir, "catalog data")); err != nil {
		t.Fatalf("custom seed directory: %v\n%s", err, output)
	}
	s.assertCounts(t, [6]int{2, 8, 8, 20, 8, 50})
}

func TestSeedTimeoutRollsBack(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "seeds")
	writeFile(t, filepath.Join(s.dir, "seeds", "99_wait.sql"), "SELECT pg_sleep(5)")
	if output, err := s.run(nil, "-timeout", "750ms"); err == nil {
		t.Fatalf("timed out seed reported success: %s", output)
	}
	s.assertCounts(t, [6]int{})
}

func TestSeedParallelRuns(t *testing.T) {
	s := newSeedDB(t, false)
	s.copySeeds(t, "seeds")
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			if output, err := s.run(nil); err != nil {
				t.Errorf("parallel seed: %v\n%s", err, output)
			}
		})
	}
	wg.Wait()
	s.assertCounts(t, [6]int{2, 8, 8, 20, 8, 50})
}
