package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/notification"
)

const FormatVersion = 1

// RestoreLockID is scoped by PostgreSQL to the connected database. Keep it stable
// across command versions so concurrent restores cannot bypass each other.
const RestoreLockID int64 = 108170741740912

type Options struct {
	DatabaseURL string
	Directory   string
	Key         []byte
	BinDir      string
}

// Manifest is published last. Its checksum detects corruption, not authenticity;
// only restore trusted local archives, which can execute database code.
type Manifest struct {
	FormatVersion    int       `json:"format_version"`
	CreatedAt        time.Time `json:"created_at"`
	SourceDatabase   string    `json:"source_database"`
	MigrationVersion int64     `json:"migration_version"`
	KeyID            string    `json:"key_id"`
	ArchiveSHA256    string    `json:"archive_sha256"`
}

// Error preserves diagnostic categories without including database credentials,
// SQL values, filesystem paths or external tool output in Error().
type Error struct {
	Stage  string
	Reason string
	cause  error
}

func (e *Error) Error() string { return e.Stage + ": " + e.Reason }
func (e *Error) Unwrap() error { return e.cause }
func fail(stage, reason string, err error) error {
	return &Error{Stage: stage, Reason: reason, cause: err}
}

func (o Options) validate() (*Connection, *notification.Cipher, error) {
	if o.Directory == "" {
		return nil, nil, fail("configuration", "backup directory is required", nil)
	}
	c, err := ParseConnection(o.DatabaseURL, os.Environ())
	if err != nil {
		return nil, nil, fail("configuration", err.Error(), nil)
	}
	cipher, err := notification.NewCipher(o.Key)
	if err != nil {
		return nil, nil, fail("configuration", "MAIL_QUEUE_KEY must contain exactly 32 bytes", nil)
	}
	return c, cipher, nil
}

func (c *Connection) connect(ctx context.Context) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(c.connectionString())
	if err != nil {
		return nil, fail("database", "connection configuration failed", err)
	}
	// ParseConfig may inspect the user's password file. Explicitly overwrite the
	// parsed password, including the intentionally empty password case.
	config.Password = c.password
	// pgx's URL parser trims every leading slash from database paths. Keep the
	// exact decoded database identity used by PGDATABASE and the manifest.
	config.Database = c.Database
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fail("database", "database connection failed", err)
	}
	return conn, nil
}

// Backup reserves a new directory and reads metadata, queue records and pg_dump
// from the same exported read-only repeatable-read transaction snapshot.
func Backup(ctx context.Context, o Options) (Manifest, error) {
	var manifest Manifest
	c, cipher, err := o.validate()
	if err != nil {
		return manifest, err
	}
	if err = ctx.Err(); err != nil {
		return manifest, fail("backup", "operation canceled", err)
	}
	if err = os.Mkdir(o.Directory, 0700); err != nil {
		return manifest, fail("directory", "cannot reserve new backup directory; it must not exist", err)
	}
	archivePath := filepath.Join(o.Directory, "database.dump")
	manifestPath := filepath.Join(o.Directory, "manifest.json")
	archiveCreated, manifestCreated, complete := false, false, false
	defer func() {
		if !complete {
			if manifestCreated {
				_ = os.Remove(manifestPath)
			}
			if archiveCreated {
				_ = os.Remove(archivePath)
			}
			_ = os.Remove(o.Directory)
		}
	}()
	conn, err := c.connect(ctx)
	if err != nil {
		return manifest, err
	}
	defer closeConn(conn)
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return manifest, fail("snapshot", "cannot start source snapshot", err)
	}
	defer rollback(tx)
	version, err := verifyState(ctx, tx, cipher)
	if err != nil {
		return manifest, err
	}
	var snapshot string
	if err = tx.QueryRow(ctx, "SELECT pg_catalog.pg_export_snapshot()").Scan(&snapshot); err != nil {
		return manifest, fail("snapshot", "cannot export source snapshot", err)
	}
	archive, err := os.OpenFile(archivePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return manifest, fail("archive", "cannot create archive", err)
	}
	archiveCreated = true
	err = runTool(ctx, o.BinDir, "pg_dump", []string{"--format=custom", "--no-owner", "--no-privileges", "--no-password", "--snapshot", snapshot}, c, nil, archive)
	if err == nil {
		err = archive.Sync()
	}
	closeErr := archive.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return manifest, fail("dump", "pg_dump failed; set POSTGRES_BIN_DIR to compatible tools and check connection settings", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return manifest, fail("snapshot", "source snapshot did not complete", err)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return manifest, fail("archive", "cannot read completed archive", err)
	}
	checksum, err := checksum(ctx, file)
	file.Close()
	if err != nil {
		return manifest, fail("archive", "cannot checksum archive", err)
	}
	manifest = Manifest{FormatVersion: FormatVersion, CreatedAt: time.Now().UTC(), SourceDatabase: c.Database, MigrationVersion: version, KeyID: cipher.KeyID(), ArchiveSHA256: checksum}
	if err = ctx.Err(); err != nil {
		return manifest, fail("backup", "operation canceled", err)
	}
	marker, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return manifest, fail("manifest", "cannot create completion manifest", err)
	}
	manifestCreated = true
	err = json.NewEncoder(marker).Encode(manifest)
	if err == nil {
		err = marker.Sync()
	}
	closeErr = marker.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return manifest, fail("manifest", "cannot finish completion manifest", err)
	}
	complete = true
	return manifest, nil
}

// Restore requires an empty separate target and holds a session advisory lock
// through pg_restore and verification. It never creates or cleans a database.
func Restore(ctx context.Context, o Options) (Manifest, error) {
	var manifest Manifest
	c, cipher, err := o.validate()
	if err != nil {
		return manifest, err
	}
	manifest, archive, err := readArchive(ctx, o.Directory)
	if err != nil {
		return manifest, err
	}
	defer archive.Close()
	if manifest.KeyID != cipher.KeyID() {
		return manifest, fail("key", "MAIL_QUEUE_KEY does not match the backup", nil)
	}
	if c.Database == manifest.SourceDatabase {
		return manifest, fail("target", "restore target must differ from source database name", nil)
	}
	conn, err := c.connect(ctx)
	if err != nil {
		return manifest, err
	}
	defer closeConn(conn)
	var locked bool
	if err = conn.QueryRow(ctx, "SELECT pg_catalog.pg_try_advisory_lock($1)", RestoreLockID).Scan(&locked); err != nil {
		return manifest, fail("lock", "cannot lock restore target", err)
	}
	if !locked {
		return manifest, fail("lock", "another restore holds the target lock", nil)
	}
	var empty bool
	if err = conn.QueryRow(ctx, emptyTargetSQL).Scan(&empty); err != nil {
		return manifest, fail("target", "cannot inspect restore target", err)
	}
	if !empty {
		return manifest, fail("target", "restore target must be completely empty", nil)
	}
	if err = runTool(ctx, o.BinDir, "pg_restore", []string{"--single-transaction", "--exit-on-error", "--no-owner", "--no-privileges", "--no-password", "--dbname", c.Database}, c, archive, io.Discard); err != nil {
		return manifest, fail("restore", "pg_restore completion was not confirmed; inspect the offline target before retrying and check POSTGRES_BIN_DIR", err)
	}
	version, err := verifyState(ctx, conn, cipher)
	if err != nil {
		return manifest, err
	}
	if version != manifest.MigrationVersion {
		return manifest, fail("verification", "restored migration version differs from manifest", nil)
	}
	return manifest, nil
}

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func verifyState(ctx context.Context, db queryer, cipher *notification.Cipher) (int64, error) {
	var version int64
	var dirty bool
	rows, err := db.Query(ctx, "SELECT version, dirty FROM public.schema_migrations")
	if err != nil {
		return 0, fail("verification", "clean schema_migrations is required", err)
	}
	count := 0
	for rows.Next() {
		count++
		if err = rows.Scan(&version, &dirty); err != nil {
			break
		}
	}
	rows.Close()
	if err == nil {
		err = rows.Err()
	}
	if err != nil || count != 1 || version <= 0 || dirty {
		return 0, fail("verification", "exactly one clean migration version is required", err)
	}
	rows, err = db.Query(ctx, "SELECT id,kind,key_id,payload,expires_at FROM public.mail_jobs WHERE expires_at > pg_catalog.transaction_timestamp()")
	if err != nil {
		return 0, fail("verification", "mail_jobs must exist and be readable", err)
	}
	defer rows.Close()
	for rows.Next() {
		var m notification.Message
		if err = rows.Scan(&m.ID, &m.Kind, &m.KeyID, &m.Ciphertext, &m.ExpiresAt); err != nil {
			return 0, fail("verification", "cannot validate queue records", err)
		}
		if _, err = cipher.Open(m); err != nil {
			return 0, fail("verification", "MAIL_QUEUE_KEY cannot decrypt unexpired queue records", nil)
		}
	}
	if err = rows.Err(); err != nil {
		return 0, fail("verification", "cannot finish queue validation", err)
	}
	return version, nil
}

const emptyTargetSQL = `SELECT NOT (
 EXISTS (SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname NOT IN ('public','pg_catalog','information_schema') AND nspname !~ '^pg_(toast|temp)(_|$)')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_extension WHERE extname <> 'plpgsql')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_type t JOIN pg_catalog.pg_namespace n ON n.oid=t.typnamespace WHERE n.nspname='public')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_proc p JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_depend d JOIN pg_catalog.pg_namespace n ON n.oid=d.refobjid WHERE d.refclassid='pg_catalog.pg_namespace'::pg_catalog.regclass AND n.nspname='public')
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_largeobject_metadata)
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_publication)
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_subscription WHERE subdbid=(SELECT oid FROM pg_catalog.pg_database WHERE datname=pg_catalog.current_database()))
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_foreign_server)
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_foreign_data_wrapper)
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_event_trigger)
 OR EXISTS (SELECT 1 FROM pg_catalog.pg_default_acl)
)`

func closeConn(c *pgx.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = c.Close(ctx)
}

func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
func checksum(ctx context.Context, r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, contextReader{ctx, r}); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil && s == strings.ToLower(s)
}
func regularFile(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return nil, fail("archive", "backup requires regular local files", nil)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fail("archive", "cannot open backup file", err)
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		file.Close()
		return nil, fail("archive", "backup file changed during open", nil)
	}
	return file, nil
}
func readArchive(ctx context.Context, dir string) (Manifest, *os.File, error) {
	var manifest Manifest
	marker, err := regularFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return manifest, nil, fail("manifest", "complete backup manifest is required", err)
	}
	stat, err := marker.Stat()
	if err != nil || stat.Size() > 64*1024 {
		marker.Close()
		return manifest, nil, fail("manifest", "invalid or oversized backup manifest", err)
	}
	decoder := json.NewDecoder(io.LimitReader(marker, 64*1024))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&manifest)
	if err == nil {
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			err = errors.New("extra manifest content")
		}
	}
	marker.Close()
	if err != nil || manifest.FormatVersion != FormatVersion || manifest.CreatedAt.IsZero() || manifest.SourceDatabase == "" || hasControl(manifest.SourceDatabase) || manifest.MigrationVersion <= 0 || !validHash(manifest.KeyID) || !validHash(manifest.ArchiveSHA256) {
		return manifest, nil, fail("manifest", "invalid or unsupported backup manifest", nil)
	}
	archive, err := regularFile(filepath.Join(dir, "database.dump"))
	if err != nil {
		return manifest, nil, err
	}
	sum, err := checksum(ctx, archive)
	if err != nil || sum != manifest.ArchiveSHA256 {
		archive.Close()
		return manifest, nil, fail("checksum", "archive checksum verification failed", err)
	}
	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		archive.Close()
		return manifest, nil, fail("archive", "cannot rewind archive", err)
	}
	return manifest, archive, nil
}

func runTool(ctx context.Context, bin, name string, args []string, c *Connection, stdin io.Reader, stdout io.Writer) error {
	program := name
	if bin != "" {
		program = filepath.Join(bin, name)
	}
	path, err := exec.LookPath(program)
	if err != nil {
		return fail("tools", "PostgreSQL tools unavailable; set POSTGRES_BIN_DIR or PATH", nil)
	}
	private, err := os.MkdirTemp("", "gotest-pgpass-")
	if err != nil {
		return fail("credentials", "cannot create private credentials directory", err)
	}
	defer os.Remove(private)
	if err = os.Chmod(private, 0700); err != nil {
		return fail("credentials", "cannot protect credentials directory", err)
	}
	passfile := filepath.Join(private, "pgpass")
	file, err := os.OpenFile(passfile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fail("credentials", "cannot create private password file", err)
	}
	defer os.Remove(passfile)
	_, err = io.WriteString(file, c.Pgpass())
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return fail("credentials", "cannot write private password file", err)
	}
	command := exec.CommandContext(ctx, path, args...)
	command.Env = append(c.ToolEnvironment(os.Environ(), passfile), "HOME="+private, "APPDATA="+private, "USERPROFILE="+private)
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = io.Discard
	if err = command.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fail("tools", "PostgreSQL tool execution failed", nil)
	}
	return nil
}
