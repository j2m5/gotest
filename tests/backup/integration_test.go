package backup_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/backup"
	"gotest/internal/notification"
	"gotest/tests/testenv"
)

type fixture struct {
	ctx       context.Context
	admin     *pgx.Conn
	adminURL  *url.URL
	source    string
	sourceURL string
	bin       string
	key       []byte
	message   notification.Message
	version   int64
}

func dbFail(t *testing.T, stage string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s failed (%T)", stage, err)
	}
}
func newFixture(t *testing.T) *fixture {
	t.Helper()
	raw := testenv.DatabaseURL(t)
	bin := os.Getenv("POSTGRES_BIN_DIR")
	for _, name := range []string{"pg_dump", "pg_restore"} {
		path := name
		if bin != "" {
			path = filepath.Join(bin, name)
		}
		if _, err := exec.LookPath(path); err != nil {
			if os.Getenv("REQUIRE_DATABASE_TESTS") == "1" || os.Getenv("CI") == "true" || os.Getenv("CI") == "1" {
				t.Fatal("PostgreSQL tools required; set POSTGRES_BIN_DIR")
			}
			t.Skip("PostgreSQL tools unavailable; set POSTGRES_BIN_DIR")
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL for isolated database tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	admin, err := pgx.Connect(ctx, raw)
	dbFail(t, "admin connection", err)
	t.Cleanup(func() { admin.Close(context.Background()) })
	f := &fixture{ctx: ctx, admin: admin, adminURL: parsed, bin: bin, key: []byte("01234567890123456789012345678901"), version: 20261027000200}
	f.source, f.sourceURL = f.database(t)
	conn := f.connect(t, f.sourceURL)
	defer conn.Close(ctx)
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	dbFail(t, "migration listing", err)
	sort.Strings(files)
	for _, file := range files {
		content, e := os.ReadFile(file)
		dbFail(t, "read migration", e)
		_, e = conn.Exec(ctx, string(content))
		dbFail(t, "apply migration", e)
	}
	_, err = conn.Exec(ctx, "CREATE TABLE schema_migrations (version bigint PRIMARY KEY, dirty boolean NOT NULL); INSERT INTO schema_migrations VALUES ($1,false)", pgx.QueryExecModeSimpleProtocol, f.version)
	dbFail(t, "migration marker", err)
	_, err = conn.Exec(ctx, "INSERT INTO users(email,login,password,name) VALUES ('backup@example.test','backup_fixture','unused','Backup')")
	dbFail(t, "fixture user", err)
	cipher, err := notification.NewCipher(f.key)
	dbFail(t, "cipher", err)
	f.message, err = cipher.Seal(notification.ResetRequest, "backup@example.test", "", time.Now().Add(time.Hour))
	dbFail(t, "seal fixture", err)
	_, err = conn.Exec(ctx, "INSERT INTO mail_jobs(id,kind,key_id,payload,expires_at) VALUES($1,$2,$3,$4,$5)", f.message.ID, string(f.message.Kind), f.message.KeyID, f.message.Ciphertext, f.message.ExpiresAt)
	dbFail(t, "fixture mail", err)
	return f
}
func (f *fixture) database(t *testing.T) (string, string) {
	t.Helper()
	var random [8]byte
	_, err := rand.Read(random[:])
	dbFail(t, "database name", err)
	name := "gotest_backup_" + hex.EncodeToString(random[:])
	_, err = f.admin.Exec(f.ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()+" TEMPLATE template0")
	dbFail(t, "create isolated database", err)
	t.Cleanup(func() {
		if !strings.HasPrefix(name, "gotest_backup_") || len(name) != 30 {
			t.Error("refusing unsafe test database cleanup")
			return
		}
		// DROP DATABASE can wait for a cluster checkpoint before removing files.
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		// A canceled DDL connection is closed asynchronously by pgx. Cleanup
		// must not make the shared CREATE DATABASE connection unusable.
		cleanup, err := pgx.Connect(ctx, f.adminURL.String())
		if err != nil {
			t.Errorf("isolated database cleanup connection failed (%T)", err)
			return
		}
		defer cleanup.Close(context.Background())
		started := time.Now()
		_, err = cleanup.Exec(ctx, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Logf("isolated database cleanup: database=%s duration=%s", name, elapsed)
		}
		if err != nil {
			t.Errorf("isolated database cleanup failed (%T)", err)
		}
	})
	u := *f.adminURL
	u.Path = "/" + name
	u.RawPath = ""
	return name, u.String()
}
func (f *fixture) connect(t *testing.T, raw string) *pgx.Conn {
	t.Helper()
	c, e := pgx.Connect(f.ctx, raw)
	dbFail(t, "isolated database connection", e)
	return c
}
func (f *fixture) options(raw, dir string) backup.Options {
	return backup.Options{DatabaseURL: raw, Directory: dir, Key: f.key, BinDir: f.bin}
}
func objectCount(t *testing.T, f *fixture, raw string) int {
	t.Helper()
	c := f.connect(t, raw)
	defer c.Close(f.ctx)
	var n int
	dbFail(t, "target objects", c.QueryRow(f.ctx, "SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public'").Scan(&n))
	return n
}

func requireTargetRefused(t *testing.T, f *fixture, raw, directory string) {
	t.Helper()
	_, err := backup.Restore(f.ctx, f.options(raw, directory))
	var problem *backup.Error
	if !errors.As(err, &problem) || problem.Stage != "target" {
		t.Fatal("nonempty target was not rejected before native restore")
	}
	if objectCount(t, f, raw) != 0 {
		t.Fatal("rejected target contains restored public objects")
	}
}

func TestBackupRestoreDrill(t *testing.T) {
	f := newFixture(t)
	dir := filepath.Join(t.TempDir(), "backup")
	manifest, err := backup.Backup(f.ctx, f.options(f.sourceURL, dir))
	dbFail(t, "backup", err)
	if manifest.SourceDatabase != f.source || manifest.MigrationVersion != f.version || manifest.KeyID != f.message.KeyID {
		t.Fatal("backup metadata did not match consistent source")
	}
	target, targetURL := f.database(t)
	_ = target
	_, err = backup.Restore(f.ctx, f.options(targetURL, dir))
	dbFail(t, "restore", err)
	conn := f.connect(t, targetURL)
	defer conn.Close(f.ctx)
	var email string
	dbFail(t, "restored user", conn.QueryRow(f.ctx, "SELECT email FROM users WHERE login='backup_fixture'").Scan(&email))
	if email != "backup@example.test" {
		t.Fatal("user row lost")
	}
	var m notification.Message
	dbFail(t, "restored message", conn.QueryRow(f.ctx, "SELECT id,kind,key_id,payload,expires_at FROM mail_jobs WHERE id=$1", f.message.ID).Scan(&m.ID, &m.Kind, &m.KeyID, &m.Ciphertext, &m.ExpiresAt))
	cipher, _ := notification.NewCipher(f.key)
	content, err := cipher.Open(m)
	dbFail(t, "restored ciphertext", err)
	if content.Email != email || m.KeyID != f.message.KeyID {
		t.Fatal("restored mail key/content changed")
	}
	var version int64
	var dirty bool
	dbFail(t, "restored version", conn.QueryRow(f.ctx, "SELECT version,dirty FROM schema_migrations").Scan(&version, &dirty))
	if dirty || version != f.version {
		t.Fatal("restored migration marker changed")
	}
	t.Run("existing directory preserved", func(t *testing.T) {
		before, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		dbFail(t, "read manifest", err)
		if _, err = backup.Backup(f.ctx, f.options(f.sourceURL, dir)); err == nil {
			t.Fatal("existing backup overwritten")
		}
		after, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if string(before) != string(after) {
			t.Fatal("existing manifest changed")
		}
	})
	t.Run("source rejected", func(t *testing.T) {
		if _, err := backup.Restore(f.ctx, f.options(f.sourceURL, dir)); err == nil {
			t.Fatal("source database accepted as restore target")
		}
	})
	t.Run("wrong key before writes", func(t *testing.T) {
		_, raw := f.database(t)
		o := f.options(raw, dir)
		o.Key = []byte("11234567890123456789012345678901")
		if _, err := backup.Restore(f.ctx, o); err == nil {
			t.Fatal("wrong queue key accepted")
		}
		if objectCount(t, f, raw) != 0 {
			t.Fatal("wrong-key restore modified target")
		}
	})
	t.Run("nonempty preserved", func(t *testing.T) {
		_, raw := f.database(t)
		c := f.connect(t, raw)
		_, err := c.Exec(f.ctx, "CREATE TABLE sentinel (value text); INSERT INTO sentinel VALUES ('preserve')")
		dbFail(t, "sentinel", err)
		c.Close(f.ctx)
		if _, err := backup.Restore(f.ctx, f.options(raw, dir)); err == nil {
			t.Fatal("nonempty target accepted")
		}
		c = f.connect(t, raw)
		defer c.Close(f.ctx)
		var value string
		dbFail(t, "sentinel read", c.QueryRow(f.ctx, "SELECT value FROM sentinel").Scan(&value))
		if value != "preserve" {
			t.Fatal("target sentinel changed")
		}
	})
	t.Run("large object target preserved", func(t *testing.T) {
		_, raw := f.database(t)
		c := f.connect(t, raw)
		defer c.Close(f.ctx)
		payload := []byte("large-object-sentinel\x00preserve")
		var oid uint32
		dbFail(t, "large object fixture", c.QueryRow(f.ctx, "SELECT pg_catalog.lo_from_bytea(0, $1)", payload).Scan(&oid))
		requireTargetRefused(t, f, raw, dir)
		var restored []byte
		dbFail(t, "large object sentinel read", c.QueryRow(f.ctx, "SELECT pg_catalog.lo_get($1)", oid).Scan(&restored))
		if !bytes.Equal(restored, payload) {
			t.Fatal("large object sentinel changed")
		}
	})
	t.Run("publication target preserved", func(t *testing.T) {
		_, raw := f.database(t)
		c := f.connect(t, raw)
		defer c.Close(f.ctx)
		_, err := c.Exec(f.ctx, "CREATE PUBLICATION backup_sentinel")
		dbFail(t, "publication fixture", err)
		requireTargetRefused(t, f, raw, dir)
		var count int
		dbFail(t, "publication sentinel read", c.QueryRow(f.ctx, "SELECT count(*) FROM pg_catalog.pg_publication WHERE pubname='backup_sentinel'").Scan(&count))
		if count != 1 {
			t.Fatal("publication sentinel changed")
		}
	})
	t.Run("global default privileges preserved", func(t *testing.T) {
		_, raw := f.database(t)
		c := f.connect(t, raw)
		defer c.Close(f.ctx)
		_, err := c.Exec(f.ctx, "ALTER DEFAULT PRIVILEGES GRANT SELECT ON TABLES TO PUBLIC")
		dbFail(t, "default privileges fixture", err)
		var before, after string
		dbFail(t, "default privileges before", c.QueryRow(f.ctx, "SELECT defaclacl::text FROM pg_catalog.pg_default_acl WHERE defaclnamespace=0 AND defaclobjtype='r'").Scan(&before))
		requireTargetRefused(t, f, raw, dir)
		dbFail(t, "default privileges after", c.QueryRow(f.ctx, "SELECT defaclacl::text FROM pg_catalog.pg_default_acl WHERE defaclnamespace=0 AND defaclobjtype='r'").Scan(&after))
		if before != after {
			t.Fatal("global default privileges changed")
		}
	})
	for _, ddl := range []string{"CREATE SCHEMA private", "CREATE TYPE public.mood AS ENUM ('happy')", "CREATE FUNCTION public.answer() RETURNS int LANGUAGE sql AS 'SELECT 42'", "CREATE EXTENSION hstore", "CREATE COLLATION public.backup_sentinel (provider = libc, locale = 'C')"} {
		t.Run("reject nonempty catalog "+strings.Fields(ddl)[1], func(t *testing.T) {
			_, raw := f.database(t)
			c := f.connect(t, raw)
			_, err := c.Exec(f.ctx, ddl)
			c.Close(f.ctx)
			dbFail(t, "catalog fixture", err)
			if _, err := backup.Restore(f.ctx, f.options(raw, dir)); err == nil {
				t.Fatal("nonempty catalog accepted")
			}
		})
	}
	t.Run("concurrent restore rejected", func(t *testing.T) {
		_, raw := f.database(t)
		c := f.connect(t, raw)
		defer c.Close(f.ctx)
		_, err := c.Exec(f.ctx, "SELECT pg_advisory_lock($1)", backup.RestoreLockID)
		dbFail(t, "restore lock fixture", err)
		if _, err := backup.Restore(f.ctx, f.options(raw, dir)); err == nil {
			t.Fatal("held restore lock ignored")
		}
		if objectCount(t, f, raw) != 0 {
			t.Fatal("locked target modified")
		}
	})
	t.Run("checksum corruption before writes", func(t *testing.T) {
		_, raw := f.database(t)
		p := filepath.Join(dir, "database.dump")
		data, err := os.ReadFile(p)
		dbFail(t, "read archive", err)
		defer os.WriteFile(p, data, 0600)
		dbFail(t, "corrupt archive", os.WriteFile(p, append(data, 0), 0600))
		if _, err := backup.Restore(f.ctx, f.options(raw, dir)); err == nil {
			t.Fatal("corrupt archive accepted")
		}
		if objectCount(t, f, raw) != 0 {
			t.Fatal("corrupt archive modified target")
		}
	})
	t.Run("invalid archive no target writes", func(t *testing.T) {
		_, raw := f.database(t)
		p := filepath.Join(dir, "database.dump")
		data, err := os.ReadFile(p)
		dbFail(t, "read archive", err)
		mp := filepath.Join(dir, "manifest.json")
		original, err := os.ReadFile(mp)
		dbFail(t, "read manifest", err)
		defer os.WriteFile(p, data, 0600)
		defer os.WriteFile(mp, original, 0600)
		invalid := []byte("not a PostgreSQL archive\n")
		sum := sha256.Sum256(invalid)
		m := manifest
		m.ArchiveSHA256 = hex.EncodeToString(sum[:])
		encoded, _ := json.Marshal(m)
		dbFail(t, "invalid archive", os.WriteFile(p, invalid, 0600))
		dbFail(t, "updated checksum", os.WriteFile(mp, encoded, 0600))
		if _, err := backup.Restore(f.ctx, f.options(raw, dir)); err == nil {
			t.Fatal("invalid archive accepted")
		}
		if objectCount(t, f, raw) != 0 {
			t.Fatal("invalid archive modified target")
		}
	})
	t.Run("canceled backup never complete", func(t *testing.T) {
		ctx, cancel := context.WithCancel(f.ctx)
		cancel()
		dir := filepath.Join(t.TempDir(), "canceled")
		if _, err := backup.Backup(ctx, f.options(f.sourceURL, dir)); err == nil {
			t.Fatal("canceled backup succeeded")
		}
		if _, err := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(err) {
			t.Fatal("canceled backup published manifest")
		}
	})
	t.Run("tool failure never complete", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "failed")
		o := f.options(f.sourceURL, dir)
		o.BinDir = t.TempDir()
		if _, err := backup.Backup(f.ctx, o); err == nil {
			t.Fatal("missing tool accepted")
		}
		if _, err := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(err) {
			t.Fatal("tool failure published manifest")
		}
	})
	t.Run("subprocess credentials and output redacted", func(t *testing.T) {
		capture := filepath.Join(t.TempDir(), "capture.json")
		t.Setenv("BACKUP_TEST_TOOL_CAPTURE", capture)
		t.Setenv("SMTP_PASSWORD", "private-smtp-secret")
		binaryDir := fakeTools(t)
		failedDir := filepath.Join(t.TempDir(), "failed-dump")
		o := f.options(f.sourceURL, failedDir)
		o.BinDir = binaryDir
		_, err := backup.Backup(f.ctx, o)
		if err == nil || strings.Contains(err.Error(), "private-") {
			t.Fatal("external tool failure succeeded or exposed private output")
		}
		assertToolCapture(t, capture, false)
		if _, err := os.Stat(failedDir); !os.IsNotExist(err) {
			t.Fatal("failed dump left its incomplete directory")
		}
		_, raw := f.database(t)
		o = f.options(raw, dir)
		o.BinDir = binaryDir
		_, err = backup.Restore(f.ctx, o)
		if err == nil || strings.Contains(err.Error(), "private-") {
			t.Fatal("restore tool failure succeeded or exposed private output")
		}
		assertToolCapture(t, capture, true)
		if objectCount(t, f, raw) != 0 {
			t.Fatal("failed restore changed target")
		}
	})
	t.Run("late archive failure rolls back all writes", func(t *testing.T) {
		c := f.connect(t, f.sourceURL)
		defer c.Close(f.ctx)
		_, err := c.Exec(f.ctx, "CREATE FUNCTION public.restore_probe() RETURNS integer LANGUAGE plpgsql AS $$ BEGIN IF current_database() <> '"+f.source+"' THEN RAISE EXCEPTION 'intentional restore drill failure'; END IF; RETURN 1; END $$; CREATE MATERIALIZED VIEW public.restore_probe_view AS SELECT public.restore_probe() AS value")
		dbFail(t, "rollback probe fixture", err)
		defer c.Exec(f.ctx, "DROP MATERIALIZED VIEW public.restore_probe_view; DROP FUNCTION public.restore_probe()")
		probeDir := filepath.Join(t.TempDir(), "rollback")
		_, err = backup.Backup(f.ctx, f.options(f.sourceURL, probeDir))
		dbFail(t, "rollback probe archive", err)
		_, raw := f.database(t)
		if _, err := backup.Restore(f.ctx, f.options(raw, probeDir)); err == nil {
			t.Fatal("late failing archive restore succeeded")
		}
		if objectCount(t, f, raw) != 0 {
			t.Fatal("late restore failure left partial schema or data")
		}
	})
	t.Run("metadata and archive share exported snapshot", func(t *testing.T) {
		capture := filepath.Join(t.TempDir(), "capture.json")
		release := filepath.Join(t.TempDir(), "continue")
		t.Setenv("BACKUP_TEST_TOOL_CAPTURE", capture)
		t.Setenv("BACKUP_TEST_TOOL_MODE", "wait-run")
		t.Setenv("BACKUP_TEST_TOOL_CONTINUE", release)
		realTool := "pg_dump"
		if f.bin != "" {
			realTool = filepath.Join(f.bin, realTool)
		}
		resolved, err := exec.LookPath(realTool)
		dbFail(t, "native tool", err)
		t.Setenv("BACKUP_TEST_REAL_TOOL", resolved)
		snapshotDir := filepath.Join(t.TempDir(), "snapshot")
		o := f.options(f.sourceURL, snapshotDir)
		o.BinDir = fakeTools(t)
		finished := make(chan error, 1)
		go func() { _, err := backup.Backup(f.ctx, o); finished <- err }()
		// Always release the wrapper and join the backup before restoring test env.
		joined := false
		defer func() {
			_ = os.WriteFile(release, []byte("go"), 0600)
			if !joined {
				<-finished
			}
		}()
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(capture); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("dump snapshot handshake timed out")
			}
			time.Sleep(10 * time.Millisecond)
		}
		c := f.connect(t, f.sourceURL)
		defer c.Close(f.ctx)
		_, err = c.Exec(f.ctx, "UPDATE users SET name='AfterSnapshot' WHERE login='backup_fixture'; UPDATE schema_migrations SET version=version+1")
		dbFail(t, "concurrent source change", err)
		defer c.Exec(f.ctx, "UPDATE users SET name='Backup' WHERE login='backup_fixture'; UPDATE schema_migrations SET version=version-1")
		dbFail(t, "release dump", os.WriteFile(release, []byte("go"), 0600))
		err = <-finished
		joined = true
		dbFail(t, "snapshot backup", err)
		t.Setenv("BACKUP_TEST_TOOL_CAPTURE", "")
		t.Setenv("BACKUP_TEST_TOOL_MODE", "")
		_, raw := f.database(t)
		_, err = backup.Restore(f.ctx, f.options(raw, snapshotDir))
		dbFail(t, "snapshot restore", err)
		restored := f.connect(t, raw)
		defer restored.Close(f.ctx)
		var name string
		dbFail(t, "snapshot user", restored.QueryRow(f.ctx, "SELECT name FROM users WHERE login='backup_fixture'").Scan(&name))
		if name != "Backup" {
			t.Fatal("archive observed data newer than its metadata snapshot")
		}
	})
	t.Run("dirty source rejected", func(t *testing.T) {
		c := f.connect(t, f.sourceURL)
		defer c.Close(f.ctx)
		_, err := c.Exec(f.ctx, "UPDATE schema_migrations SET dirty=true")
		dbFail(t, "dirty fixture", err)
		defer c.Exec(f.ctx, "UPDATE schema_migrations SET dirty=false")
		dir := filepath.Join(t.TempDir(), "dirty")
		if _, err := backup.Backup(f.ctx, f.options(f.sourceURL, dir)); err == nil {
			t.Fatal("dirty source backed up")
		}
		if _, err := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(err) {
			t.Fatal("dirty backup published manifest")
		}
	})
	t.Run("source queue wrong key rejected", func(t *testing.T) {
		o := f.options(f.sourceURL, filepath.Join(t.TempDir(), "wrong-key"))
		o.Key = []byte("11234567890123456789012345678901")
		if _, err := backup.Backup(f.ctx, o); err == nil {
			t.Fatal("source key mismatch ignored")
		}
	})
	if objectCount(t, f, f.sourceURL) == 0 {
		t.Fatal("source database damaged")
	}
}
