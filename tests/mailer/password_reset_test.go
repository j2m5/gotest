package mailer_test

import (
	"context"
	"errors"
	"gotest/internal/mailer"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readOnlyMessage(t *testing.T, path string) (mail.Header, string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	message, err := mail.ReadMessage(file)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(quotedprintable.NewReader(message.Body))
	if err != nil {
		t.Fatal(err)
	}
	return message.Header, string(body)
}

func TestPasswordResetMailUsesSMTPTransport(t *testing.T) {
	address, result := fakeSMTP(t, "starttls", trustedCertificate)
	m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: address, SMTPUsername: "user", SMTPPassword: "pass", SMTPTLSMode: "starttls"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SendPasswordReset(context.Background(), "user@example.com", "a+b&c"); err != nil {
		t.Fatal(err)
	}
	got := <-result
	if got.err != nil {
		t.Fatal(got.err)
	}
	if !strings.Contains(got.message, "https://example.com/reset-password?token=3Da%2Bb%26c") || !strings.Contains(got.message, "If you did not request a password reset") {
		t.Fatalf("incorrect reset message: %s", got.message)
	}
}

func TestPasswordResetMailContainsEncodedSingleUseLinkAndIgnoreText(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SendPasswordReset(context.Background(), "user@example.com", "a+b&c"); err != nil {
		t.Fatal(err)
	}
	entries, err := filepath.Glob(filepath.Join(o.Directory, "*.eml"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries=%v err=%v", entries, err)
	}
	h, body := readOnlyMessage(t, entries[0])
	if h.Get("Subject") != "Reset your password" {
		t.Fatalf("subject %q", h.Get("Subject"))
	}
	if !strings.Contains(body, "https://example.com/reset-password?token=a%2Bb%26c") || !strings.Contains(strings.ToLower(body), "single-use") || !strings.Contains(strings.ToLower(body), "if you did not request") {
		t.Fatalf("incomplete reset email: %s", body)
	}
}

func TestPasswordChangedMailHasNoCredentialsAndClosesSessions(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SendPasswordChanged(context.Background(), "user@example.com"); err != nil {
		t.Fatal(err)
	}
	entries, err := filepath.Glob(filepath.Join(o.Directory, "*.eml"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries=%v err=%v", entries, err)
	}
	h, body := readOnlyMessage(t, entries[0])
	if h.Get("Subject") != "Your password was changed" {
		t.Fatalf("subject %q", h.Get("Subject"))
	}
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "all sessions have been closed") || !strings.Contains(lower, "sign in again") || strings.Contains(lower, "token=") || strings.Contains(lower, "password:") {
		t.Fatalf("unsafe or incomplete password change email: %s", body)
	}
}

func TestPasswordRecoveryMailRejectsInvalidInputsAndCancelledContext(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SendPasswordReset(context.Background(), "user@example.com", ""); err == nil {
		t.Fatal("accepted empty token")
	}
	if err := m.SendPasswordReset(context.Background(), "bad", "token"); err == nil {
		t.Fatal("accepted invalid recipient")
	}
	if err := m.SendPasswordChanged(context.Background(), "bad"); err == nil {
		t.Fatal("accepted invalid recipient")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.SendPasswordReset(ctx, "user@example.com", "token"); !errors.Is(err, context.Canceled) {
		t.Fatalf("reset returned %v", err)
	}
	if err := m.SendPasswordChanged(ctx, "user@example.com"); !errors.Is(err, context.Canceled) {
		t.Fatalf("changed returned %v", err)
	}
	entries, _ := os.ReadDir(o.Directory)
	if len(entries) != 0 {
		t.Fatal("wrote rejected email")
	}
}
