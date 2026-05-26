package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	AppPort     string
	AppEnv      string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:123@localhost:5432/gotest"),
		AppPort:     getEnv("APP_PORT", "8081"),
		AppEnv:      getEnv("APP_ENV", "local"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)

	if val == "" {
		if fallback == "" {
			log.Fatalf("%s environment variable not set", key)
		}

		return fallback
	}

	return val
}
