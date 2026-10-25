package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gotest/internal/mailer"
)

type Config struct {
	DatabaseURL, AppPort, AppEnv, PublicURL string
	SessionTTL, VerificationTTL             time.Duration
	Mail                                    mailer.Options
}

// Load reads configuration without opening connections or terminating the process.
func Load() (*Config, error) {
	cfg := &Config{DatabaseURL: os.Getenv("DATABASE_URL"), AppPort: getEnv("APP_PORT", "8081"), AppEnv: getEnv("APP_ENV", "local")}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	port, err := strconv.Atoi(cfg.AppPort)
	if err != nil || port < 1 || port > 65535 || strings.IndexFunc(cfg.AppPort, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return nil, errors.New("APP_PORT must be an integer between 1 and 65535")
	}
	cfg.PublicURL = getEnv("PUBLIC_URL", "http://localhost:"+cfg.AppPort)
	public, err := url.Parse(cfg.PublicURL)
	if err != nil || public == nil || (public.Scheme != "http" && public.Scheme != "https") || public.Hostname() == "" || public.User != nil || (public.Path != "" && public.Path != "/") || public.RawQuery != "" || public.ForceQuery || public.Fragment != "" || public.RawFragment != "" || strings.Contains(cfg.PublicURL, "#") {
		return nil, errors.New("PUBLIC_URL must be an HTTP or HTTPS origin")
	}
	if cfg.AppEnv == "production" && public.Scheme != "https" {
		return nil, errors.New("PUBLIC_URL must use HTTPS in production")
	}
	// Reuse the validated, normalized origin for links, CSRF trust and cookie security.
	cfg.PublicURL = public.Scheme + "://" + strings.ToLower(public.Host)
	cfg.SessionTTL, err = durationEnv("SESSION_TTL", 168*time.Hour, 8760*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.VerificationTTL, err = durationEnv("VERIFICATION_TTL", 24*time.Hour, 8760*time.Hour)
	if err != nil {
		return nil, err
	}
	timeout, err := durationEnv("SMTP_TIMEOUT", 10*time.Second, time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.Mail = mailer.Options{Driver: getEnv("MAIL_DRIVER", "local"), From: getEnv("MAIL_FROM", "noreply@example.test"), PublicURL: cfg.PublicURL, Directory: getEnv("MAIL_DIR", ".local/mail"), SMTPAddress: os.Getenv("SMTP_ADDR"), SMTPUsername: os.Getenv("SMTP_USERNAME"), SMTPPassword: os.Getenv("SMTP_PASSWORD"), SMTPTLSMode: getEnv("SMTP_TLS_MODE", "starttls"), Timeout: timeout}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback, maximum time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 || duration > maximum {
		return 0, fmt.Errorf("%s must be a positive duration no greater than %s", key, maximum)
	}
	return duration, nil
}
