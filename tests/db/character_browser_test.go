package db_test

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gotest/internal/auth"
)

// This optional smoke test uses the real built SPA, HTTP API and isolated SQL
// schema. Only the Chromium binary is external; no browser package is required.
func TestCharacterBrowserFlow(t *testing.T) {
	browser := os.Getenv("CHARACTER_BROWSER_BINARY")
	if browser == "" {
		t.Skip("set CHARACTER_BROWSER_BINARY and build frontend to run the browser smoke test")
	}
	if _, err := os.Stat(filepath.Join("..", "..", "frontend", "dist", "index.html")); err != nil {
		t.Fatal("build frontend before the browser smoke test")
	}
	s, ctx := characterTestStorage(t, true)
	mail := &flowMail{}
	h, service := characterFlowRouter(t, s, mail)
	if err := service.Register(ctx, auth.RegisterInput{Email: "browser@example.test", Login: "browser", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, mail.token); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h)
	defer server.Close()
	browserCtx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	command := exec.CommandContext(browserCtx, "node", filepath.Join("..", "frontend", "characters.browser.mjs"), server.URL, browser, t.TempDir())
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("browser smoke test: %v\n%s", err, output)
	}
	t.Log(string(output))
}
