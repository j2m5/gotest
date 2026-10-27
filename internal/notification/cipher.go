package notification

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"
)

var errInvalidMessage = errors.New("invalid notification message")

type Cipher struct {
	aead  cipher.AEAD
	keyID string
}

func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, errors.New("notification key must contain 32 bytes")
	}
	block, err := aes.NewCipher(append([]byte(nil), key...))
	if err != nil {
		return nil, errors.New("notification cipher initialization failed")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("notification cipher initialization failed")
	}
	sum := sha256.Sum256(key)
	return &Cipher{aead: aead, keyID: hex.EncodeToString(sum[:])}, nil
}
func (c *Cipher) KeyID() string { return c.keyID }

func (c *Cipher) Seal(kind Kind, email, token string, expires time.Time) (Message, error) {
	content := Content{Email: email, Token: token}
	if !validContent(kind, content) || expires.IsZero() {
		return Message{}, errInvalidMessage
	}
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return Message{}, errors.New("notification randomness failed")
	}
	m := Message{ID: hex.EncodeToString(id), Kind: kind, KeyID: c.keyID, ExpiresAt: expires.UTC().Truncate(time.Microsecond)}
	payload, err := json.Marshal(content)
	if err != nil {
		return Message{}, errInvalidMessage
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return Message{}, errors.New("notification randomness failed")
	}
	m.Ciphertext = c.aead.Seal(nonce, nonce, payload, messageAAD(m))
	return m, nil
}

func (c *Cipher) Open(m Message) (Content, error) {
	if !validHex(m.ID) || m.KeyID != c.keyID || !validKind(m.Kind) || m.ExpiresAt.IsZero() || len(m.Ciphertext) < c.aead.NonceSize()+c.aead.Overhead() {
		return Content{}, errInvalidMessage
	}
	n := c.aead.NonceSize()
	plaintext, err := c.aead.Open(nil, m.Ciphertext[:n], m.Ciphertext[n:], messageAAD(m))
	if err != nil {
		return Content{}, errInvalidMessage
	}
	var content Content
	if err := json.Unmarshal(plaintext, &content); err != nil || !validContent(m.Kind, content) {
		return Content{}, errInvalidMessage
	}
	return content, nil
}
func messageAAD(m Message) []byte {
	data, _ := json.Marshal(struct {
		ID     string
		Kind   Kind
		KeyID  string
		Expiry string
	}{m.ID, m.Kind, m.KeyID, m.ExpiresAt.UTC().Format(time.RFC3339Nano)})
	return data
}
func validHex(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func validKind(k Kind) bool {
	return k == Verification || k == PasswordReset || k == PasswordChanged || k == ResetRequest
}
func validContent(kind Kind, c Content) bool {
	if !validKind(kind) || c.Email == "" || len(c.Email) > 254 || strings.IndexFunc(c.Email, unicode.IsControl) >= 0 {
		return false
	}
	address, err := mail.ParseAddress(c.Email)
	if err != nil || address.Name != "" || address.Address != c.Email || !strings.Contains(c.Email, "@") {
		return false
	}
	if kind == Verification || kind == PasswordReset {
		return validHex(c.Token)
	}
	return c.Token == ""
}
