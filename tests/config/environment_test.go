package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gotest/internal/config"
)

func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "configuration with spaces")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(directory, "app.env")
	if err := os.WriteFile(name, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func allowFileValues(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadFileResolvesPathsAndDoesNotMutateEnvironment(t *testing.T) {
	clearConfigEnv(t)
	allowFileValues(t, "APP_PORT", "STATIC_DIR", "MAIL_DIR", "MAIL_DRIVER", "MAIL_FROM", "SMTP_ADDR")
	file := writeConfigFile(t, "APP_PORT=9090\nSTATIC_DIR=web assets\nMAIL_DIR=mail output\nMAIL_DRIVER=smtp\nMAIL_FROM=sender@example.test\nSMTP_ADDR=smtp.example.test:587\n")
	t.Chdir(t.TempDir())
	cfg, err := config.LoadWithOptions(config.Options{EnvFile: file})
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Dir(file)
	if cfg.AppPort != "9090" || cfg.PublicURL != "http://localhost:9090" || cfg.StaticDir != filepath.Join(base, "web assets") || cfg.Mail.Directory != filepath.Join(base, "mail output") {
		t.Fatal("file values or config-relative paths were ignored")
	}
	if cfg.Mail.Driver != "smtp" || cfg.Mail.From != "sender@example.test" || cfg.Mail.SMTPAddress != "smtp.example.test:587" {
		t.Fatal("file mail settings not loaded")
	}
	for _, key := range []string{"APP_PORT", "STATIC_DIR", "MAIL_DIR", "MAIL_DRIVER", "MAIL_FROM", "SMTP_ADDR"} {
		if _, exists := os.LookupEnv(key); exists {
			t.Fatalf("loading config mutated %s", key)
		}
	}
	for _, directory := range []string{cfg.StaticDir, cfg.Mail.Directory} {
		if _, err := os.Stat(directory); !os.IsNotExist(err) {
			t.Fatal("config loading created a data directory")
		}
	}
}

func TestLoadFileEnvironmentOverridesIncludeEmptyValues(t *testing.T) {
	clearConfigEnv(t)
	file := writeConfigFile(t, "APP_PORT=9090\nSTATIC_DIR=file-assets\nMAIL_DIR=file-mail\nMAIL_DRIVER=smtp\n")
	mail := filepath.Join(t.TempDir(), "absolute mail")
	t.Setenv("APP_PORT", "8181")
	t.Setenv("MAIL_DIR", mail)
	// Explicitly empty environment values mask file settings and use documented defaults.
	t.Setenv("STATIC_DIR", "")
	t.Setenv("MAIL_DRIVER", "")
	cfg, err := config.LoadWithOptions(config.Options{EnvFile: file})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppPort != "8181" || cfg.Mail.Driver != "local" || cfg.Mail.Directory != mail || cfg.StaticDir != filepath.Join(filepath.Dir(file), "frontend", "dist") {
		t.Fatal("environment precedence/default paths changed")
	}
	assets := filepath.Join(t.TempDir(), "absolute assets")
	t.Setenv("STATIC_DIR", assets)
	cfg, err = config.LoadWithOptions(config.Options{EnvFile: file})
	if err != nil || cfg.StaticDir != assets {
		t.Fatalf("absolute override: %v", err)
	}
	t.Setenv("STATIC_DIR", "process assets")
	t.Setenv("MAIL_DIR", "process mail")
	cfg, err = config.LoadWithOptions(config.Options{EnvFile: file})
	if err != nil || cfg.StaticDir != filepath.Join(filepath.Dir(file), "process assets") || cfg.Mail.Directory != filepath.Join(filepath.Dir(file), "process mail") {
		t.Fatalf("relative process overrides changed the selected base: %v", err)
	}
}

func TestLoadFileOptionalMissingAndSafeFailures(t *testing.T) {
	clearConfigEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)
	missing := filepath.Join(t.TempDir(), "private-missing.env")
	if _, err := config.LoadWithOptions(config.Options{EnvFile: missing}); err == nil || strings.Contains(err.Error(), "private-") {
		t.Fatal("required file error missing or leaked its path")
	}
	cfg, err := config.LoadWithOptions(config.Options{EnvFile: missing, OptionalEnvFile: true})
	if err != nil || cfg.StaticDir != filepath.Join(cwd, "frontend", "dist") || cfg.Mail.Directory != filepath.Join(cwd, ".local", "mail") {
		t.Fatalf("optional missing file = %v", err)
	}
	malformed := writeConfigFile(t, "SMTP_PASSWORD=\"private-secret-without-closing-quote\n")
	for _, file := range []string{malformed, filepath.Dir(malformed)} {
		if _, err := config.LoadWithOptions(config.Options{EnvFile: file, OptionalEnvFile: true}); err == nil || strings.Contains(err.Error(), "private-") {
			t.Fatal("unreadable/malformed optional file was ignored or leaked its value")
		}
	}
	// Environment-only loading must not read an unrelated dotenv file in cwd.
	if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("invalid key = \"private-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(); err != nil {
		t.Fatalf("environment-only load read cwd file: %v", err)
	}
}

func TestLoadEnvironmentPathsUseStartupDirectory(t *testing.T) {
	clearConfigEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("STATIC_DIR", filepath.Join("site", "..", "web assets"))
	t.Setenv("MAIL_DIR", "mail output")
	cfg, err := config.Load()
	if err != nil || cfg.StaticDir != filepath.Join(cwd, "web assets") || cfg.Mail.Directory != filepath.Join(cwd, "mail output") {
		t.Fatalf("cwd-relative paths = %v", err)
	}
	for _, key := range []string{"STATIC_DIR", "MAIL_DIR"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "   ")
			if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("blank path accepted for %s: %v", key, err)
			}
		})
		t.Run(key+" NUL from file", func(t *testing.T) {
			allowFileValues(t, key)
			file := writeConfigFile(t, key+"=private-\x00-directory\n")
			if _, err := config.LoadWithOptions(config.Options{EnvFile: file}); err == nil || strings.Contains(err.Error(), "private-") {
				t.Fatal("NUL path accepted or exposed")
			}
		})
	}
	if runtime.GOOS == "windows" {
		unc := `\\files.example.test\share\web assets`
		t.Setenv("STATIC_DIR", unc)
		cfg, err := config.Load()
		if err != nil || cfg.StaticDir != unc {
			t.Fatalf("absolute UNC path was changed: %v", err)
		}
		for _, value := range []string{`C:private-dir`, `\private-dir`} {
			t.Setenv("STATIC_DIR", value)
			if _, err := config.Load(); err == nil || strings.Contains(err.Error(), "private-") {
				t.Fatal("ambiguous Windows path accepted or exposed")
			}
		}
	}
}
