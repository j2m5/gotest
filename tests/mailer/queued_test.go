package mailer_test

import (
	"context"
	"gotest/internal/mailer"
	"gotest/internal/notification"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueuedDeliveryHasStableMessageIDDateAndToken(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	d := notification.Delivery{ID: strings.Repeat("a", 64), Kind: notification.PasswordReset, Email: "user@example.com", Token: strings.Repeat("b", 64), CreatedAt: time.Date(2026, 10, 27, 10, 11, 12, 0, time.UTC)}
	for range 2 {
		if err := m.Deliver(context.Background(), d); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(o.Directory)
	if err != nil || len(entries) != 2 {
		t.Fatalf("files: %d %v", len(entries), err)
	}
	for _, entry := range entries {
		f, err := os.Open(filepath.Join(o.Directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		msg, err := mail.ReadMessage(f)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if msg.Header.Get("Message-ID") != "<"+strings.Repeat("a", 64)+"@example.com>" || msg.Header.Get("Date") != "Tue, 27 Oct 2026 10:11:12 +0000" {
			t.Fatal("unstable headers")
		}
		if !strings.Contains(string(body), "https://example.com/reset-password?token="+strings.Repeat("b", 64)) {
			t.Fatal("changed link")
		}
	}
}
func TestQueuedDeliveryKindsAndValidation(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []notification.Kind{notification.Verification, notification.PasswordChanged} {
		d := notification.Delivery{ID: strings.Repeat("c", 64), Kind: kind, Email: "user@example.com", CreatedAt: time.Now()}
		if kind == notification.Verification {
			d.Token = strings.Repeat("d", 64)
		}
		if err := m.Deliver(context.Background(), d); err != nil {
			t.Fatal(err)
		}
	}
	for _, change := range []func(*notification.Delivery){func(d *notification.Delivery) { d.ID = "bad\r\nBcc: user@example.com" }, func(d *notification.Delivery) { d.Kind = notification.ResetRequest }, func(d *notification.Delivery) { d.Email = "a@example.com\nBcc: b@example.com" }, func(d *notification.Delivery) { d.CreatedAt = time.Time{} }, func(d *notification.Delivery) { d.Token = "bad" }} {
		d := notification.Delivery{ID: strings.Repeat("c", 64), Kind: notification.Verification, Email: "user@example.com", Token: strings.Repeat("d", 64), CreatedAt: time.Now()}
		change(&d)
		if err := m.Deliver(context.Background(), d); err == nil {
			t.Fatal("accepted unsafe delivery")
		}
	}
	entries, _ := os.ReadDir(o.Directory)
	if len(entries) != 2 {
		t.Fatal("invalid message written")
	}
}
