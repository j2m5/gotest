package db_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCleanupCommand(t *testing.T) {
	s, ctx := cleanupFixture(t)
	insertExpired(t, ctx, s, 3)
	binary := filepath.Join(t.TempDir(), "cleanup.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/cleanup")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build cleanup: %v\n%s", err, output)
	}
	directory := t.TempDir()
	run := func(dsn string, args ...string) ([]byte, error) {
		runContext, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		command := exec.CommandContext(runContext, binary, args...)
		command.Dir = directory
		for _, variable := range os.Environ() {
			key, _, _ := strings.Cut(variable, "=")
			if strings.EqualFold(key, "DATABASE_URL") || strings.EqualFold(key, "APP_PORT") || strings.EqualFold(key, "MAIL_DRIVER") {
				continue
			}
			command.Env = append(command.Env, variable)
		}
		// HTTP/mail configuration must not be required for a maintenance command.
		command.Env = append(command.Env, "DATABASE_URL="+dsn, "APP_PORT=invalid", "MAIL_DRIVER=invalid")
		return command.CombinedOutput()
	}
	readEvent := func(t *testing.T, output []byte, event string) map[string]any {
		t.Helper()
		decoder := json.NewDecoder(bytes.NewReader(output))
		for {
			var record map[string]any
			if err := decoder.Decode(&record); err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("cleanup output is not structured JSON: %v", err)
			}
			if record["msg"] == event {
				return record
			}
		}
		t.Fatalf("missing event %s", event)
		return nil
	}
	dsn := s.pool.Config().ConnString()
	// ConnString retains the originally parsed DSN, not later RuntimeParams edits.
	// Carry the fixture's isolated schema into the separate CLI process explicitly.
	schema := s.pool.Config().ConnConfig.RuntimeParams["search_path"]
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("cannot construct isolated CLI test connection")
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		dsn = parsed.String()
	} else {
		dsn += " search_path=" + schema
	}
	t.Run("preview", func(t *testing.T) {
		output, err := run(dsn, "-dry-run", "-batch-size", "2")
		if err != nil {
			event := readEvent(t, output, "cleanup_failed")
			t.Fatalf("preview failed: %v stage=%v sqlstate=%v", err, event["stage"], event["sqlstate"])
		}
		event := readEvent(t, output, "cleanup_preview")
		if event["sessions_matched"] != float64(3) || event["verifications_matched"] != float64(3) {
			t.Fatal("incorrect preview counts")
		}
		cleanupCounts(t, ctx, s, 3, 3)
	})
	t.Run("delete", func(t *testing.T) {
		output, err := run(dsn, "-batch-size", "2", "-timeout", "5s")
		if err != nil {
			event := readEvent(t, output, "cleanup_failed")
			t.Fatalf("cleanup failed: %v stage=%v sqlstate=%v", err, event["stage"], event["sqlstate"])
		}
		event := readEvent(t, output, "cleanup_completed")
		if event["sessions_deleted"] != float64(3) || event["verifications_deleted"] != float64(3) || event["batches"] != float64(4) {
			t.Fatal("incorrect committed deletion counts")
		}
		cleanupCounts(t, ctx, s, 0, 0)
	})
	for _, args := range [][]string{{"-batch-size", "0"}, {"-batch-size", "10001"}, {"-timeout", "0"}, {"unexpected"}, {"-batch-size", "private-argument"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := run("postgres://private-user:private-password@%zz/test", args...)
			if err == nil {
				t.Fatal("invalid arguments succeeded")
			}
			if strings.Contains(string(output), "private-") {
				t.Fatal("argument values or credentials logged")
			}
			if readEvent(t, output, "cleanup_failed")["stage"] != "arguments" {
				t.Fatal("invalid flags reached database connection")
			}
		})
	}
	t.Run("database failure", func(t *testing.T) {
		output, err := run("postgres://private-user:private-password@%zz/test")
		if err == nil || strings.Contains(string(output), "private-") {
			t.Fatal("database error succeeded or disclosed credentials")
		}
		if readEvent(t, output, "cleanup_failed")["stage"] != "database" {
			t.Fatal("database failure stage missing")
		}
	})
	t.Run("help without configuration", func(t *testing.T) {
		output, err := run("", "-help")
		if err != nil || !strings.Contains(string(output), "dry-run") {
			t.Fatal("help requires database configuration")
		}
	})
}
