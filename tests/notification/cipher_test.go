package notification_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"gotest/internal/notification"
	"strings"
	"testing"
	"time"
)

func newCipher(t *testing.T) *notification.Cipher {
	t.Helper()
	c, err := notification.NewCipher(bytes.Repeat([]byte{42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestCipherRoundTripPrecisionAndImmutableKey(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	c, err := notification.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(key)
	key[0] = 0
	if c.KeyID() != hex.EncodeToString(sum[:]) {
		t.Fatal("key ID changed")
	}
	expiry := time.Date(2027, 1, 2, 3, 4, 5, 123456789, time.FixedZone("offset", 3600))
	m, err := c.Seal(notification.Verification, "user@example.com", strings.Repeat("a", 64), expiry)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ID) != 64 || m.ExpiresAt.Location() != time.UTC || m.ExpiresAt.Nanosecond() != 123456000 {
		t.Fatal("metadata not normalized")
	}
	if bytes.Contains(m.Ciphertext, []byte("user@example.com")) || bytes.Contains(m.Ciphertext, []byte(strings.Repeat("a", 64))) {
		t.Fatal("plaintext persisted")
	}
	content, err := c.Open(m)
	if err != nil || content.Email != "user@example.com" || content.Token != strings.Repeat("a", 64) {
		t.Fatalf("roundtrip failed: %v", err)
	}
	second, err := c.Seal(notification.Verification, content.Email, content.Token, expiry)
	if err != nil || second.ID == m.ID || bytes.Equal(second.Ciphertext, m.Ciphertext) {
		t.Fatal("randomness missing")
	}
}
func TestCipherRejectsTamperingAndOtherKey(t *testing.T) {
	c := newCipher(t)
	m, err := c.Seal(notification.PasswordReset, "user@example.com", strings.Repeat("b", 64), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for name, alter := range map[string]func(*notification.Message){
		"id": func(m *notification.Message) { m.ID = strings.Repeat("c", 64) }, "kind": func(m *notification.Message) { m.Kind = notification.Verification }, "key": func(m *notification.Message) { m.KeyID = strings.Repeat("d", 64) }, "expiry": func(m *notification.Message) { m.ExpiresAt = m.ExpiresAt.Add(time.Microsecond) }, "payload": func(m *notification.Message) {
			m.Ciphertext = append([]byte(nil), m.Ciphertext...)
			m.Ciphertext[len(m.Ciphertext)-1] ^= 1
		}, "short": func(m *notification.Message) { m.Ciphertext = []byte{1} },
	} {
		t.Run(name, func(t *testing.T) {
			copy := m
			alter(&copy)
			if _, err := c.Open(copy); err == nil {
				t.Fatal("accepted altered message")
			}
		})
	}
	other, _ := notification.NewCipher(bytes.Repeat([]byte{43}, 32))
	if _, err := other.Open(m); err == nil {
		t.Fatal("accepted other key")
	}
}
func TestCipherRejectsInvalidContentWithoutLeaking(t *testing.T) {
	for _, n := range []int{0, 31, 33} {
		if _, err := notification.NewCipher(make([]byte, n)); err == nil {
			t.Fatal("invalid key accepted")
		}
	}
	c := newCipher(t)
	for _, v := range []struct {
		kind         notification.Kind
		email, token string
	}{
		{"unknown", "user@example.com", ""}, {notification.Verification, "user@example.com", "secret"}, {notification.Verification, "user@example.com", strings.Repeat("A", 64)}, {notification.ResetRequest, "user@example.com", "secret"}, {notification.PasswordChanged, "user@example.com", "secret"}, {notification.ResetRequest, "user@example.com\r\nBcc:secret", ""}, {notification.ResetRequest, "Name <user@example.com>", ""},
	} {
		_, err := c.Seal(v.kind, v.email, v.token, time.Now().Add(time.Hour))
		if err == nil {
			t.Fatal("invalid content accepted")
		}
		if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), v.email) {
			t.Fatal("unsafe error")
		}
	}
	for _, kind := range []notification.Kind{notification.ResetRequest, notification.PasswordChanged} {
		m, err := c.Seal(kind, "user@example.com", "", time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = c.Open(m); err != nil {
			t.Fatal(err)
		}
	}
}
