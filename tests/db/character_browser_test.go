package db_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/tests/testenv"
)

// This optional smoke test uses the real built SPA, HTTP API and isolated SQL
// schema. Only the Chromium binary is external; no browser package is required.
func TestCharacterBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
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
	if err := service.Register(ctx, auth.RegisterInput{Email: "unverified@example.test", Login: "unverified", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Chromium forbids overriding Sec-Fetch-Site through extra headers.
		// Inject that metadata at the test boundary so the real origin guard
		// produces a readable 403 for the login UI, without disabling CORS.
		if r.URL.Path == "/api/auth/login" && r.Header.Get("X-Browser-Test-Cross-Site") == "1" {
			r.Header.Set("Sec-Fetch-Site", "cross-site")
			r.Header.Set("Origin", "https://untrusted.example")
		}
		h.ServeHTTP(w, r)
	}))
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
