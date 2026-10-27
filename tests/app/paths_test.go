package app_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func appTestEnvironment(extra ...string) []string {
	var env []string
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		key = strings.ToUpper(key)
		if key == "DATABASE_URL" || strings.HasPrefix(key, "APP_") || key == "PUBLIC_URL" || key == "STATIC_DIR" || key == "TRUSTED_PROXIES" || key == "SESSION_TTL" || key == "REMEMBER_ME_TTL" || key == "PASSWORD_RESET_TTL" || key == "VERIFICATION_TTL" || strings.HasPrefix(key, "MAIL_") || strings.HasPrefix(key, "SMTP_") {
			continue
		}
		env = append(env, variable)
	}
	// Public fixture key; never used for application data.
	env = append(env, "MAIL_QUEUE_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	return append(env, extra...)
}

func TestAppArgumentsAndExplicitEnvironmentFile(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "app.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/app")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build app: %v\n%s", err, output)
	}
	selected := filepath.Join(t.TempDir(), "selected configuration.env")
	// SMTP settings come only from the file; constructing the transport does not send mail.
	if err := os.WriteFile(selected, []byte("DATABASE_URL=postgres://private-user:private-password@%zz/test\nMAIL_DRIVER=smtp\nMAIL_FROM=sender@example.test\nSMTP_ADDR=smtp.example.test:587\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name      string
		args, env []string
		stage     string
		help      bool
	}{
		{"help", []string{"-help"}, nil, "", true},
		{"unknown argument", []string{"-private-argument"}, nil, "arguments", false},
		{"positional argument", []string{"private-argument"}, nil, "arguments", false},
		{"missing flag value", []string{"-env-file"}, nil, "arguments", false},
		{"missing explicit file", []string{"-env-file", filepath.Join(t.TempDir(), "private-missing.env")}, []string{"DATABASE_URL=postgres://private-user:private-password@%zz/test"}, "configuration", false},
		{"malformed implicit file", nil, []string{"DATABASE_URL=postgres://private-user:private-password@%zz/test"}, "configuration", false},
		{"file selection and SMTP sender", []string{"-env-file", selected}, nil, "database", false},
		{"disabled dotenv", []string{"-env-file="}, []string{"DATABASE_URL=postgres://private-user:private-password@%zz/test"}, "database", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cwd := t.TempDir()
			if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("SMTP_PASSWORD=\"private-malformed-secret\n"), 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(ctx, binary, tt.args...)
			command.Dir, command.Env = cwd, appTestEnvironment(tt.env...)
			output, err := command.CombinedOutput()
			if strings.Contains(string(output), "private-") {
				t.Fatal("CLI/config values exposed in output")
			}
			if tt.help {
				if err != nil || !strings.Contains(string(output), "-env-file") {
					t.Fatal("help requires valid environment or does not describe file selection")
				}
				return
			}
			if err == nil {
				t.Fatal("invalid startup succeeded")
			}
			found := false
			for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				var entry map[string]any
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					t.Fatal("startup failure output is not JSON")
				}
				if entry["msg"] == "application_failed" && entry["stage"] == tt.stage {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing expected failure stage %s", tt.stage)
			}
		})
	}
}
