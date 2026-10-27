package config_test

import (
	"gotest/internal/config"
	"strings"
	"testing"
	"time"
)

func TestRememberMeTTLValidation(t *testing.T) {
	for _, value := range []string{"invalid", "0s", "-1h", "8761h", "1h"} {
		t.Run(value, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("SESSION_TTL", "24h")
			t.Setenv("REMEMBER_ME_TTL", value)
			if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "REMEMBER_ME_TTL") {
				t.Fatalf("invalid remember lifetime accepted: %v", err)
			}
		})
	}
}

func TestRememberMeTTLFileAndEnvironment(t *testing.T) {
	clearConfigEnv(t)
	allowFileValues(t, "SESSION_TTL", "REMEMBER_ME_TTL")
	file := writeConfigFile(t, "SESSION_TTL=2h\nREMEMBER_ME_TTL=240h\n")
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"file", 240 * time.Hour}, {"48h", 48 * time.Hour}, {"2h", 2 * time.Hour}, {"", 720 * time.Hour},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if tc.value != "file" {
				t.Setenv("REMEMBER_ME_TTL", tc.value)
			}
			cfg, err := config.LoadWithOptions(config.Options{EnvFile: file})
			if err != nil {
				t.Fatal(err)
			}
			if cfg.SessionTTL != 2*time.Hour || cfg.RememberMeTTL != tc.want {
				t.Fatalf("lifetimes: session=%v remember=%v", cfg.SessionTTL, cfg.RememberMeTTL)
			}
		})
	}
}
