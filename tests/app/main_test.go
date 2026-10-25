package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartupFailuresUseSafeStructuredDiagnostics(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "app.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/app")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build app: %v\n%s", err, output)
	}
	for _, tc := range []struct{ name, port, dsn, proxies, stage string }{
		{"configuration", "private-port", "postgres://unused", "", "configuration"},
		{"proxies", "8081", "postgres://unused", "private-proxy.example.test", "configuration"},
		{"database", "8081", "postgres://private-user:private-password@%zz/test", "", "database"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := exec.CommandContext(ctx, binary)
			command.Dir = t.TempDir()
			for _, variable := range os.Environ() {
				key, _, _ := strings.Cut(variable, "=")
				key = strings.ToUpper(key)
				if key == "DATABASE_URL" || strings.HasPrefix(key, "APP_") || key == "PUBLIC_URL" || key == "TRUSTED_PROXIES" || key == "SESSION_TTL" || key == "VERIFICATION_TTL" || strings.HasPrefix(key, "MAIL_") || strings.HasPrefix(key, "SMTP_") {
					continue
				}
				command.Env = append(command.Env, variable)
			}
			command.Env = append(command.Env, "DATABASE_URL="+tc.dsn, "APP_PORT="+tc.port, "TRUSTED_PROXIES="+tc.proxies, "MAIL_DRIVER=local")
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatal("invalid startup configuration succeeded")
			}
			if strings.Contains(string(output), "private-") {
				t.Fatal("configuration values leaked into startup output")
			}
			decoder := json.NewDecoder(bytes.NewReader(output))
			found := false
			for {
				var entry map[string]any
				if err := decoder.Decode(&entry); err == io.EOF {
					break
				} else if err != nil {
					t.Fatalf("startup output is not JSON: %v", err)
				}
				if entry["msg"] == "application_failed" && entry["level"] == "ERROR" && entry["stage"] == tc.stage && entry["error_kind"] != nil {
					found = true
				}
			}
			if !found {
				t.Fatal("missing structured failure stage and category")
			}
		})
	}
}
