package config_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"gotest/internal/config"
)

const publicQueueKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func TestMailQueueKeyRequiresBase64EncodingOf32Bytes(t *testing.T) {
	for name, value := range map[string]string{
		"missing": "", "invalid base64": "private-sensitive-value!", "short": base64.StdEncoding.EncodeToString([]byte("private-sensitive-value")), "31 bytes": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{31}, 31)), "33 bytes": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{33}, 33)),
	} {
		t.Run(name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("MAIL_QUEUE_KEY", value)
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), "MAIL_QUEUE_KEY") {
				t.Fatal("invalid queue key accepted or error omitted setting name")
			}
			if strings.Contains(err.Error(), "private-sensitive-value") || (value != "" && strings.Contains(err.Error(), value)) {
				t.Fatal("queue key leaked in validation error")
			}
		})
	}
}

func TestMailQueueKeyDecodesAndIsExcludedFromJSON(t *testing.T) {
	clearConfigEnv(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(cfg.MailQueueKey, []byte("0123456789abcdef0123456789abcdef")) {
		t.Fatal("decoded key changed")
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"MailQueueKey", publicQueueKey, "0123456789abcdef0123456789abcdef"} {
		if strings.Contains(string(data), secret) {
			t.Fatal("queue key exposed in JSON")
		}
	}
}

func TestMailQueueKeyFileAndProcessPrecedence(t *testing.T) {
	for _, mode := range []string{"file", "process", "explicit empty masks file"} {
		t.Run(mode, func(t *testing.T) {
			clearConfigEnv(t)
			file := writeConfigFile(t, "MAIL_QUEUE_KEY="+publicQueueKey+"\n")
			expected := []byte("0123456789abcdef0123456789abcdef")
			switch mode {
			case "file":
				allowFileValues(t, "MAIL_QUEUE_KEY")
			case "process":
				expected = bytes.Repeat([]byte{77}, 32)
				t.Setenv("MAIL_QUEUE_KEY", base64.StdEncoding.EncodeToString(expected))
			case "explicit empty masks file":
				t.Setenv("MAIL_QUEUE_KEY", "")
			}
			cfg, err := config.LoadWithOptions(config.Options{EnvFile: file})
			if mode == "explicit empty masks file" {
				if err == nil || !strings.Contains(err.Error(), "MAIL_QUEUE_KEY") {
					t.Fatal("empty process value did not mask file key")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(cfg.MailQueueKey, expected) {
				t.Fatal("key source precedence changed")
			}
		})
	}
}

func TestEarlierConfigErrorPrecedesMissingMailQueueKey(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("MAIL_QUEUE_KEY", "")
	t.Setenv("APP_PORT", "invalid")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "APP_PORT") {
		t.Fatal("queue key validation masked earlier config error")
	}
}
