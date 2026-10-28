package db_test

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueueStatusCommand(t *testing.T) {
	s, ctx, q := queueStorage(t)
	if err := q.EnqueueResetRequest(ctx, queueMessage(1)); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "mailqueue.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/mailqueue")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build mailqueue: %v\n%s", err, output)
	}
	directory := t.TempDir()
	run := func(dsn string, args ...string) ([]byte, error) {
		deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		command := exec.CommandContext(deadline, binary, args...)
		command.Dir = directory
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if strings.EqualFold(key, "DATABASE_URL") || strings.EqualFold(key, "MAIL_DRIVER") || strings.EqualFold(key, "MAIL_QUEUE_KEY") {
				continue
			}
			command.Env = append(command.Env, entry)
		}
		command.Env = append(command.Env, "DATABASE_URL="+dsn, "MAIL_DRIVER=private-invalid-driver", "MAIL_QUEUE_KEY=private-invalid-key")
		return command.CombinedOutput()
	}
	event := func(t *testing.T, output []byte, name string) map[string]any {
		t.Helper()
		var record map[string]any
		if err := json.Unmarshal(output, &record); err != nil {
			t.Fatal("status must emit one JSON event")
		}
		if record["msg"] != name {
			t.Fatalf("unexpected status event %v stage=%v", record["msg"], record["stage"])
		}
		return record
	}
	dsn := s.pool.Config().ConnString()
	schema := s.pool.Config().ConnConfig.RuntimeParams["search_path"]
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("cannot construct isolated CLI connection")
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		dsn = parsed.String()
	} else {
		dsn += " search_path=" + schema
	}
	// A conflicting .env proves the process environment wins from another cwd.
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte("DATABASE_URL=private-invalid-default\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Run("status uses only database configuration", func(t *testing.T) {
		output, err := run(dsn, "-timeout", "5s")
		if err != nil {
			t.Fatalf("status failed: %v stage=%v", err, event(t, output, "mail_queue_status_failed")["stage"])
		}
		r := event(t, output, "mail_queue_status")
		if r["total"] != float64(1) || r["pending"] != float64(1) || r["due"] != float64(1) || r["leased"] != float64(0) || r["failed"] != float64(0) || r["expired"] != float64(0) || r["cutoff"] == nil || r["oldest_pending_age_seconds"] == nil {
			t.Fatal("incorrect aggregate fields")
		}
		var count int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM mail_jobs`).Scan(&count); err != nil || count != 1 {
			t.Fatal("status changed queue")
		}
	})
	for _, args := range [][]string{{"-timeout", "0"}, {"-timeout", "-1s"}, {"-timeout", "private-argument"}, {"-private-flag"}, {"private-positional"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := run("postgres://private-user:private-password@%zz/private-db", args...)
			if err == nil || strings.Contains(string(output), "private-") {
				t.Fatal("flags accepted or private arguments logged")
			}
			if event(t, output, "mail_queue_status_failed")["stage"] != "arguments" {
				t.Fatal("flag failure reached configuration")
			}
		})
	}
	t.Run("bad connection redacts credentials", func(t *testing.T) {
		for _, bad := range []string{"postgres://private-user:private-password@%zz/private-db", "postgres://private-user:private-password@127.0.0.1:1/private-db?sslmode=disable"} {
			output, err := run(bad, "-timeout", "1s")
			if err == nil || strings.Contains(string(output), "private-") || strings.Contains(string(output), "postgres://") {
				t.Fatal("database error accepted or disclosed configuration")
			}
			if event(t, output, "mail_queue_status_failed")["stage"] != "database" {
				t.Fatal("wrong database failure stage")
			}
		}
	})
	t.Run("help without configuration", func(t *testing.T) {
		output, err := run("", "-help")
		if err != nil || !strings.Contains(string(output), "Usage: mailqueue") {
			t.Fatal("help requires configuration")
		}
	})
	t.Run("missing configuration", func(t *testing.T) {
		output, err := run("")
		if err == nil || event(t, output, "mail_queue_status_failed")["stage"] != "configuration" {
			t.Fatal("missing configuration accepted")
		}
	})
}
