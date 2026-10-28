package mailer_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gotest/internal/mailer"
	"gotest/internal/notification"
)

func TestLocalTextPreviewPreservesCopyableLinks(t *testing.T) {
	for _, queued := range []bool{false, true} {
		for _, kind := range []notification.Kind{notification.Verification, notification.PasswordReset} {
			name := string(kind)
			if queued {
				name += "/queued"
			}
			t.Run(name, func(t *testing.T) {
				o := localOptions(t)
				o.PublicURL = "http://localhost:5173"
				m, err := mailer.New(o)
				if err != nil {
					t.Fatal(err)
				}
				token := strings.Repeat("abcdef01", 8)
				if queued {
					err = m.Deliver(context.Background(), notification.Delivery{ID: strings.Repeat("a", 64), Kind: kind, Email: "user@example.com", Token: token, CreatedAt: time.Now()})
				} else if kind == notification.Verification {
					err = m.SendVerification(context.Background(), "user@example.com", token)
				} else {
					err = m.SendPasswordReset(context.Background(), "user@example.com", token)
				}
				if err != nil {
					t.Fatal(err)
				}
				files, err := filepath.Glob(filepath.Join(o.Directory, "*.eml"))
				if err != nil || len(files) != 1 {
					t.Fatalf("MIME messages = %d, error = %v", len(files), err)
				}
				preview := strings.TrimSuffix(files[0], ".eml") + ".txt"
				content, err := os.ReadFile(preview)
				if err != nil {
					t.Fatalf("readable local mail preview is missing: %v", err)
				}
				route := "/verify"
				if kind == notification.PasswordReset {
					route = "/reset-password"
				}
				// A user copying a line from the text file must get the complete URL,
				// without needing a MIME decoder or joining quoted-printable lines.
				want := "http://localhost:5173" + route + "?token=" + token
				found := false
				for _, line := range strings.Split(string(content), "\n") {
					if strings.TrimSuffix(line, "\r") == want {
						found = true
					}
				}
				if !found || !strings.Contains(string(content), "To: user@example.com") {
					t.Fatal("preview lost the recipient or the copyable link")
				}
				info, err := os.Stat(preview)
				if err != nil {
					t.Fatal(err)
				}
				if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
					t.Fatalf("preview permissions = %v", info.Mode())
				}
			})
		}
	}
}
