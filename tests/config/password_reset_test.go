package config_test

import (
	"gotest/internal/config"
	"testing"
	"time"
)

func TestPasswordResetTTLRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"0s", "-1m", "25h", "invalid"} {
		t.Run(value, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("PASSWORD_RESET_TTL", value)
			if _, err := config.Load(); err == nil {
				t.Fatal("invalid reset lifetime accepted")
			}
		})
	}
}

func TestPasswordResetTTLFileAndEnvironment(t *testing.T) {
	clearConfigEnv(t)
	allowFileValues(t, "PASSWORD_RESET_TTL")
	file := writeConfigFile(t, "PASSWORD_RESET_TTL=10m\n")
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"file", 10 * time.Minute}, {"24h", 24 * time.Hour}, {"", 30 * time.Minute},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if tc.value != "file" {
				t.Setenv("PASSWORD_RESET_TTL", tc.value)
			}
			cfg, err := config.LoadWithOptions(config.Options{EnvFile: file})
			if err != nil {
				t.Fatal(err)
			}
			if cfg.PasswordResetTTL != tc.want {
				t.Fatalf("reset TTL=%v want%v", cfg.PasswordResetTTL, tc.want)
			}
		})
	}
}
