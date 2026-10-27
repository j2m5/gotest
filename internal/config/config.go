package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gotest/internal/clientip"
	"gotest/internal/mailer"
)

type Config struct {
	DatabaseURL, AppPort, AppEnv, PublicURL    string
	StaticDir                                  string
	SessionTTL, RememberMeTTL, VerificationTTL time.Duration
	PasswordResetTTL                           time.Duration
	Mail                                       mailer.Options
	TrustedProxies                             clientip.Policy
}

// Options selects a dotenv source. Empty EnvFile reads only process variables.
type Options struct {
	EnvFile         string
	OptionalEnvFile bool
}

// Load reads process configuration without opening connections or changing the environment.
func Load() (*Config, error) {
	return LoadWithOptions(Options{})
}

// LoadWithOptions merges the optional file with process variables and resolves
// data paths against the file directory (or startup cwd when no file was loaded).
func LoadWithOptions(options Options) (*Config, error) {
	value, base, err := loadEnvironment(options)
	if err != nil {
		return nil, err
	}
	getEnv := func(key, fallback string) string {
		if v := value(key); v != "" {
			return v
		}
		return fallback
	}
	cfg := &Config{DatabaseURL: value("DATABASE_URL"), AppPort: getEnv("APP_PORT", "8081"), AppEnv: getEnv("APP_ENV", "local")}
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
	cfg.TrustedProxies, err = clientip.Parse(value("TRUSTED_PROXIES"))
	if err != nil {
		return nil, err
	}
	cfg.SessionTTL, err = durationEnv(value, "SESSION_TTL", 24*time.Hour, 8760*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.RememberMeTTL, err = durationEnv(value, "REMEMBER_ME_TTL", 720*time.Hour, 8760*time.Hour)
	if err != nil {
		return nil, err
	}
	if cfg.RememberMeTTL < cfg.SessionTTL {
		return nil, errors.New("REMEMBER_ME_TTL must be at least SESSION_TTL")
	}
	cfg.VerificationTTL, err = durationEnv(value, "VERIFICATION_TTL", 24*time.Hour, 8760*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.PasswordResetTTL, err = durationEnv(value, "PASSWORD_RESET_TTL", 30*time.Minute, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	timeout, err := durationEnv(value, "SMTP_TIMEOUT", 10*time.Second, time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.StaticDir, err = resolveDirectory(base, getEnv("STATIC_DIR", "frontend/dist"), "STATIC_DIR")
	if err != nil {
		return nil, err
	}
	mailDir, err := resolveDirectory(base, getEnv("MAIL_DIR", ".local/mail"), "MAIL_DIR")
	if err != nil {
		return nil, err
	}
	cfg.Mail = mailer.Options{Driver: getEnv("MAIL_DRIVER", "local"), From: getEnv("MAIL_FROM", "noreply@example.test"), PublicURL: cfg.PublicURL, Directory: mailDir, SMTPAddress: value("SMTP_ADDR"), SMTPUsername: value("SMTP_USERNAME"), SMTPPassword: value("SMTP_PASSWORD"), SMTPTLSMode: getEnv("SMTP_TLS_MODE", "starttls"), Timeout: timeout}
	return cfg, nil
}

func durationEnv(value func(string) string, key string, fallback, maximum time.Duration) (time.Duration, error) {
	raw := value(key)
	if raw == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 || duration > maximum {
		return 0, fmt.Errorf("%s must be a positive duration no greater than %s", key, maximum)
	}
	return duration, nil
}
