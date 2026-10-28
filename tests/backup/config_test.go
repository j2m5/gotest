package backup_test

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"

	"gotest/internal/backup"
)

func TestConnectionRejectsAmbiguousAndUnsafeInputs(t *testing.T) {
	bad := []string{
		"host=localhost dbname=test", "postgres://localhost/test", "postgres://user@/test",
		"postgres://user@localhost/", "postgres://user@localhost/a%3Db", "postgres://user@localhost/postgres%3A%2F%2Ffoo",
		"postgres://user@host,other/test", "postgres://user@host/test#fragment",
		"postgres://user@host/test?sslmode=disable&sslmode=require", "postgres://user@host/test?service=secret",
		"postgres://user@host/test?host=other", "postgres://user@host/test?search_path=public",
		"postgres://user:secret%0Aline@host/test", "postgres://user@host/test?sslmode=invalid",
		"postgres://user@host:0/test", "postgres://user@host/test?connect_timeout=-1",
		"postgres://user@host/test?sslmode=verify-full", "postgres://user@host/test?sslmode=verify-ca",
		"postgres://user@host/test?sslrootcert=system&sslmode=require",
		"postgres://user@host/test?sslcrl=private", "postgres://user@host/test?sslcrldir=private",
		"postgres://user@host/test?ssl_min_protocol_version=TLSv1.2", "postgres://user@host/test?ssl_max_protocol_version=TLSv1.3",
	}
	for _, raw := range bad {
		if _, err := backup.ParseConnection(raw, nil); err == nil {
			t.Errorf("unsafe URL case accepted: %d", len(raw))
		} else if strings.Contains(err.Error(), "secret") {
			t.Fatal("configuration error contains input")
		}
	}
	for _, key := range []string{"PGHOST", "PGPASSWORD", "PGSERVICE", "PGOPTIONS", "PGSSLMODE", "pgdatabase"} {
		if _, err := backup.ParseConnection("postgres://user@host/test", []string{key + "=private"}); err == nil {
			t.Errorf("ambient %s accepted", key)
		}
	}
	if _, err := backup.ParseConnection("postgres://user@host/test", []string{"PGHOST=", "POSTGRES_BIN_DIR=private"}); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionToolEnvironmentAndPgpass(t *testing.T) {
	c, err := backup.ParseConnection("postgres://a%3Ab:p%3A%5Cword@[::1]:5433/test?sslmode=disable&connect_timeout=3&channel_binding=disable", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Host != "::1" || c.Port != "5433" || c.User != "a:b" || c.Database != "test" {
		t.Fatal("connection identity lost")
	}
	if got := c.Pgpass(); got != "\\:\\:1:5433:test:a\\:b:p\\:\\\\word\n" {
		t.Fatalf("pgpass escaping mismatch: %q", got)
	}
	env := c.ToolEnvironment([]string{"PATH=tools", "PGPASSWORD=private", "PGOPTIONS=private", "DATABASE_URL=private", "RESTORE_DATABASE_URL=private", "TEST_DATABASE_URL=private", "MAIL_QUEUE_KEY=private", "SMTP_PASSWORD=private"}, "passfile")
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "private") || strings.Contains(joined, "PGPASSWORD=") {
		t.Fatal("sensitive ambient environment inherited")
	}
	for _, setting := range []string{"PATH=tools", "PGHOST=::1", "PGPORT=5433", "PGUSER=a:b", "PGDATABASE=test", "PGPASSFILE=passfile", "PGSSLMODE=disable", "PGCONNECT_TIMEOUT=3", "PGCHANNELBINDING=disable"} {
		if !strings.Contains(joined, setting) {
			t.Errorf("lost tool option %s", setting)
		}
	}
	for _, key := range []string{"PGSSLROOTCERT", "PGSSLCERT", "PGSSLKEY", "PGSSLCRL"} {
		if !strings.Contains(joined, key+"="+filepath.Join(".", "unused-"+strings.ToLower(key))) {
			t.Errorf("ambient TLS file fallback remains enabled for %s", key)
		}
	}
}

func TestConnectionPreservesSupportedTLSOptions(t *testing.T) {
	c, err := backup.ParseConnection("postgres://user@host/test?sslmode=verify-full&sslrootcert=ca.pem&sslcert=client.pem&sslkey=client.key", nil)
	if err != nil {
		t.Fatal(err)
	}
	env := strings.Join(c.ToolEnvironment(nil, "passfile"), "\n")
	for _, value := range []string{"PGSSLMODE=verify-full", "PGSSLROOTCERT=ca.pem", "PGSSLCERT=client.pem", "PGSSLKEY=client.key"} {
		if !strings.Contains(env, value) {
			t.Fatal("explicit TLS parameter was lost")
		}
	}
	c, err = backup.ParseConnection("postgres://user@host/test?sslrootcert=system", nil)
	if err != nil || !strings.Contains(strings.Join(c.ToolEnvironment(nil, "passfile"), "\n"), "PGSSLMODE=verify-full") {
		t.Fatal("system trust did not require full verification")
	}
}

func TestQueueKeyRequiresExactly32Base64Bytes(t *testing.T) {
	for _, raw := range []string{"", "private", base64.StdEncoding.EncodeToString(make([]byte, 31)), base64.StdEncoding.EncodeToString(make([]byte, 33))} {
		if _, err := backup.ParseKey(raw); err == nil {
			t.Fatal("invalid queue key accepted")
		}
	}
	if key, err := backup.ParseKey(base64.StdEncoding.EncodeToString(make([]byte, 32))); err != nil || len(key) != 32 {
		t.Fatal("valid queue key rejected")
	}
}
