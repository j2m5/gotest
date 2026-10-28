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
	"runtime"
	"strings"
	"testing"
)

func localOptions(t *testing.T) mailer.Options {
	t.Helper()
	return mailer.Options{Driver: "local", From: "sender@example.com", PublicURL: "https://example.com/base/", Directory: filepath.Join(t.TempDir(), "mail")}
}

func TestLocalCreatesSeparatePrivateMIMEFiles(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = m.SendVerification(context.Background(), "user@example.com", "a+b&c"); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(o.Directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("got %d files, want two MIME/text pairs", len(entries))
	}
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if extension != ".eml" && extension != ".txt" {
			t.Fatalf("wrong extension: %s", entry.Name())
		}
		info, _ := entry.Info()
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
			t.Fatalf("permissions: %v", info.Mode())
		}
		if extension == ".txt" {
			continue
		}
		if _, err := os.Stat(filepath.Join(o.Directory, strings.TrimSuffix(entry.Name(), ".eml")+".txt")); err != nil {
			t.Fatal("missing paired text preview")
		}
		f, err := os.Open(filepath.Join(o.Directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		msg, err := mail.ReadMessage(f)
		if err != nil {
			t.Fatal(err)
		}
		if msg.Header.Get("From") != "sender@example.com" || msg.Header.Get("To") != "user@example.com" {
			t.Fatal("bad envelope headers")
		}
		if msg.Header.Get("MIME-Version") != "1.0" || msg.Header.Get("Content-Type") != "text/plain; charset=UTF-8" {
			t.Fatal("missing MIME headers")
		}
		body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "https://example.com/verify?token=a%2Bb%26c") {
			t.Fatalf("bad URL: %s", body)
		}
	}
}

func TestRejectsUnsafeConfiguration(t *testing.T) {
	for _, change := range []func(*mailer.Options){
		func(o *mailer.Options) { o.Driver = "unknown" }, func(o *mailer.Options) { o.From = "a@example.com\r\nBcc: b@example.com" },
		func(o *mailer.Options) { o.From = "Name <a@example.com>" }, func(o *mailer.Options) { o.PublicURL = "javascript:alert(1)" },
		func(o *mailer.Options) { o.PublicURL = "https://user:pass@example.com" }, func(o *mailer.Options) { o.PublicURL = "https://example.com/?evil=x" },
		func(o *mailer.Options) { o.Driver = "smtp"; o.SMTPAddress = "localhost:25"; o.SMTPTLSMode = "none" },
	} {
		o := localOptions(t)
		change(&o)
		if _, err := mailer.New(o); err == nil {
			t.Fatalf("accepted unsafe mailer.Options: %+v", o)
		}
	}
}

func TestRejectsRecipientInjectionAndCancelledContext(t *testing.T) {
	o := localOptions(t)
	m, err := mailer.New(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, email := range []string{"a@example.com\r\nBcc: b@example.com", "Name <a@example.com>", "bad", ""} {
		if err := m.SendVerification(context.Background(), email, "token"); err == nil {
			t.Fatalf("accepted %q", email)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.SendVerification(ctx, "a@example.com", "token"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	entries, _ := os.ReadDir(o.Directory)
	if len(entries) != 0 {
		t.Fatal("wrote rejected email")
	}
}
