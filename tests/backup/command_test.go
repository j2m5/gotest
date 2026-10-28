package backup_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommandSafety(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "dbbackup.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/dbbackup")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("dbbackup build failed: %v %s", err, output)
	}
	base := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "PG") || upper == "DATABASE_URL" || upper == "RESTORE_DATABASE_URL" || upper == "MAIL_QUEUE_KEY" {
			continue
		}
		base = append(base, entry)
	}
	base = append(base, "DATABASE_URL=postgres://private-user:private-password@host/app", "MAIL_QUEUE_KEY="+base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")))
	directory := t.TempDir()
	run := func(args ...string) ([]byte, error) {
		c := exec.CommandContext(ctx, binary, args...)
		c.Dir = directory
		c.Env = base
		return c.CombinedOutput()
	}
	for _, args := range [][]string{{"-help"}, {"backup", "-help"}, {"restore", "-help"}} {
		output, err := run(args...)
		if err != nil || !strings.Contains(string(output), "Usage:") {
			t.Fatal("help requires environment or database")
		}
	}
	for _, args := range [][]string{{}, {"private-command"}, {"backup", "-dir", "private-path", "-timeout", "private-argument"}, {"restore", "-dir", "private-path", "-timeout", "0"}, {"backup", "-dir", "private-path", "extra-private"}, {"backup"}} {
		output, err := run(args...)
		if err == nil {
			t.Fatal("invalid arguments accepted")
		}
		if strings.Contains(string(output), "private-") {
			t.Fatal("CLI echoed arbitrary argument or credential")
		}
		var event map[string]any
		if json.Unmarshal(output, &event) != nil || event["msg"] != "dbbackup_failed" || event["stage"] != "arguments" {
			t.Fatalf("argument failure was not controlled JSON: %s", output)
		}
	}
	output, err := run("restore", "-dir", filepath.Join(directory, "backup"))
	if err == nil || strings.Contains(string(output), "private-") {
		t.Fatal("missing explicit restore URL accepted or secret disclosed")
	}
	var event map[string]any
	if json.Unmarshal(output, &event) != nil || event["stage"] != "configuration" || event["setting"] != "RESTORE_DATABASE_URL" {
		t.Fatal("restore did not require explicit target")
	}
	base = append(base, "RESTORE_DATABASE_URL=postgres://private-user:private-password@%zz/test")
	output, err = run("restore", "-dir", filepath.Join(directory, "backup"))
	if err == nil || strings.Contains(string(output), "private-") {
		t.Fatal("invalid URL accepted or credentials disclosed")
	}
	if json.Unmarshal(output, &event) != nil || event["stage"] != "configuration" {
		t.Fatal("invalid URL did not fail safely")
	}
}
