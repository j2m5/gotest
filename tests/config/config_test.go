package config_test

import (
	"gotest/internal/config"
	"testing"
	"time"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"DATABASE_URL", "APP_PORT", "APP_ENV", "PUBLIC_URL", "SESSION_TTL", "VERIFICATION_TTL", "MAIL_DRIVER", "MAIL_FROM", "MAIL_DIR", "SMTP_ADDR", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_TLS_MODE", "SMTP_TIMEOUT"} {
		t.Setenv(key, "")
	}
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://test:test@localhost/test" || cfg.AppPort != "8081" || cfg.AppEnv != "local" || cfg.PublicURL != "http://localhost:8081" {
		t.Fatalf("incorrect application defaults: %+v", cfg)
	}
	if cfg.SessionTTL != 168*time.Hour || cfg.VerificationTTL != 24*time.Hour {
		t.Fatal("incorrect auth lifetimes")
	}
	if cfg.Mail.Driver != "local" || cfg.Mail.From != "noreply@example.test" || cfg.Mail.Directory != ".local/mail" || cfg.Mail.PublicURL != cfg.PublicURL || cfg.Mail.SMTPTLSMode != "starttls" || cfg.Mail.Timeout != 10*time.Second {
		t.Fatalf("incorrect mail defaults: %+v", cfg.Mail)
	}
}

func TestLoadMissingDatabaseReturnsError(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("missing database accepted")
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	for _, tt := range []struct{ key, value string }{
		{"APP_PORT", "0"}, {"APP_PORT", "65536"}, {"APP_PORT", "abc"}, {"APP_PORT", "+80"},
		{"PUBLIC_URL", "javascript:alert(1)"}, {"PUBLIC_URL", "https://user:pass@example.com"}, {"PUBLIC_URL", "https://example.com/path"}, {"PUBLIC_URL", "https://example.com/?q=x"}, {"PUBLIC_URL", "https://example.com/#fragment"}, {"PUBLIC_URL", "https://"},
		{"SESSION_TTL", "0s"}, {"SESSION_TTL", "-1h"}, {"SESSION_TTL", "8761h"}, {"VERIFICATION_TTL", "invalid"}, {"VERIFICATION_TTL", "-1s"},
		{"SMTP_TIMEOUT", "0s"}, {"SMTP_TIMEOUT", "61s"}, {"SMTP_TIMEOUT", "invalid"},
	} {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv(tt.key, tt.value)
			if _, err := config.Load(); err == nil {
				t.Fatal("invalid setting accepted")
			}
		})
	}
}

func TestLoadProductionRequiresHTTPS(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "production")
	if _, err := config.Load(); err == nil {
		t.Fatal("production HTTP accepted")
	}
	t.Setenv("PUBLIC_URL", "https://example.com/")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PublicURL != "https://example.com" {
		t.Fatalf("unexpected URL %q", cfg.PublicURL)
	}
}

func TestLoadMapsSettingsWithoutRequiringSMTP(t *testing.T) {
	clearConfigEnv(t)
	values := map[string]string{"APP_PORT": "9090", "APP_ENV": "staging", "PUBLIC_URL": "https://example.com", "SESSION_TTL": "2h", "VERIFICATION_TTL": "30m", "MAIL_DRIVER": "smtp", "MAIL_FROM": "configured@example.com", "MAIL_DIR": "custom-mail", "SMTP_ADDR": "smtp.example.com:465", "SMTP_USERNAME": "test-user", "SMTP_PASSWORD": "test-password", "SMTP_TLS_MODE": "tls", "SMTP_TIMEOUT": "45s"}
	for key, value := range values {
		t.Setenv(key, value)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppPort != "9090" || cfg.AppEnv != "staging" || cfg.SessionTTL != 2*time.Hour || cfg.VerificationTTL != 30*time.Minute {
		t.Fatal("application settings not mapped")
	}
	m := cfg.Mail
	if m.Driver != "smtp" || m.From != "configured@example.com" || m.Directory != "custom-mail" || m.PublicURL != "https://example.com" || m.SMTPAddress != "smtp.example.com:465" || m.SMTPUsername != "test-user" || m.SMTPPassword != "test-password" || m.SMTPTLSMode != "tls" || m.Timeout != 45*time.Second {
		t.Fatal("mail settings not mapped")
	}
	t.Setenv("SMTP_ADDR", "")
	if _, err := config.Load(); err != nil {
		t.Fatalf("seed requires only config, not SMTP construction: %v", err)
	}
}

func TestLoadPublicURLUsesConfiguredPort(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_PORT", "9000")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PublicURL != "http://localhost:9000" {
		t.Fatal("public URL ignored configured port")
	}
}
