package db_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

type pathApp struct {
	done         chan struct{}
	err          error
	url, logFile string
}

func pathAppEnvironment(values map[string]string) []string {
	var env []string
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		key = strings.ToUpper(key)
		if key == "DATABASE_URL" || strings.HasPrefix(key, "APP_") || key == "PUBLIC_URL" || key == "STATIC_DIR" || key == "TRUSTED_PROXIES" || key == "SESSION_TTL" || key == "REMEMBER_ME_TTL" || key == "PASSWORD_RESET_TTL" || key == "VERIFICATION_TTL" || strings.HasPrefix(key, "MAIL_") || strings.HasPrefix(key, "SMTP_") {
			continue
		}
		env = append(env, variable)
	}
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	env = append(env, "MAIL_QUEUE_KEY="+publicMailQueueKey)
	return env
}

func pathAppPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	_ = listener.Close()
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func startPathApp(t *testing.T, ctx context.Context, binary, cwd, port string, values map[string]string, args ...string) *pathApp {
	t.Helper()
	logFile := filepath.Join(t.TempDir(), "app.log")
	log, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir, command.Env, command.Stdout, command.Stderr = cwd, pathAppEnvironment(values), log, log
	if err := command.Start(); err != nil {
		_ = log.Close()
		t.Fatal(err)
	}
	a := &pathApp{done: make(chan struct{}), url: "http://127.0.0.1:" + port, logFile: logFile}
	go func() { a.err = command.Wait(); close(a.done) }()
	t.Cleanup(func() { _ = command.Process.Kill(); <-a.done; _ = log.Close() })
	return a
}

func waitPathApp(t *testing.T, ctx context.Context, a *pathApp) {
	t.Helper()
	client := &http.Client{Timeout: 300 * time.Millisecond}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		response, err := client.Get(a.url + "/readyz")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == 200 {
				return
			}
		}
		select {
		case <-a.done:
			t.Fatalf("app exited before readiness: %v", a.err)
		case <-ctx.Done():
			t.Fatal("app readiness timed out")
		case <-ticker.C:
		}
	}
}

func checkPathAppResponse(t *testing.T, a *pathApp, method, path, body string, status int, fragment string) {
	t.Helper()
	req, err := http.NewRequest(method, a.url+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<16))
	if err != nil || response.StatusCode != status || !strings.Contains(string(data), fragment) {
		t.Fatalf("%s %s: status=%d, expected body match=%t, err=%v", method, path, response.StatusCode, strings.Contains(string(data), fragment), err)
	}
	if method == "HEAD" && len(data) != 0 {
		t.Fatal("HEAD returned a body")
	}
}

func TestAppPathsFromExternalWorkingDirectory(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	binary := filepath.Join(t.TempDir(), "app.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/app")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build app: %v\n%s", err, output)
	}
	// Carry the isolated fixture schema explicitly across the process boundary.
	dsn := s.pool.Config().ConnString()
	schema := s.pool.Config().ConnConfig.RuntimeParams["search_path"]
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("cannot construct isolated app connection")
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		dsn = parsed.String()
	} else {
		dsn += " search_path=" + schema
	}
	configurationDir := filepath.Join(t.TempDir(), "deployment with spaces")
	assets := filepath.Join(configurationDir, "web assets")
	if err := os.MkdirAll(filepath.Join(assets, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "index.html"), []byte("<html>external configuration</html>"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "assets", "app.js"), []byte("globalThis.pathProbe=true"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Run("explicit file anchors assets and local mail", func(t *testing.T) {
		port, cwd := pathAppPort(t), t.TempDir()
		// This broken unrelated file must never be consulted.
		if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("PRIVATE=\"unclosed\n"), 0600); err != nil {
			t.Fatal(err)
		}
		values := map[string]string{"DATABASE_URL": dsn, "APP_PORT": port, "APP_ENV": "production", "PUBLIC_URL": "https://game.example.test", "STATIC_DIR": "web assets", "MAIL_DRIVER": "local", "MAIL_DIR": "outgoing letters"}
		contents, err := godotenv.Marshal(values)
		if err != nil {
			t.Fatal("cannot encode isolated test configuration")
		}
		file := filepath.Join(configurationDir, "application.env")
		if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		a := startPathApp(t, ctx, binary, cwd, port, nil, "-env-file", file)
		waitPathApp(t, ctx, a)
		checkPathAppResponse(t, a, "GET", "/healthz", "", 200, "ok")
		checkPathAppResponse(t, a, "GET", "/characters/new", "", 200, "external configuration")
		checkPathAppResponse(t, a, "GET", "/assets/app.js", "", 200, "pathProbe")
		checkPathAppResponse(t, a, "HEAD", "/assets/app.js", "", 200, "")
		checkPathAppResponse(t, a, "GET", "/assets/missing.js", "", 404, "404")
		checkPathAppResponse(t, a, "GET", "/api/auth/me", "", 401, "unauthenticated")
		checkPathAppResponse(t, a, "GET", "/api/missing", "", 404, "route_not_found")
		checkPathAppResponse(t, a, "POST", "/api/auth/register", `{"email":"paths@example.test","login":"paths_user","password":"test-password-123"}`, 201, `"verification_queued":true`)
		var letters []string
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			letters, err = filepath.Glob(filepath.Join(configurationDir, "outgoing letters", "*.eml"))
			if err != nil || len(letters) > 0 {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if err != nil || len(letters) != 1 {
			t.Fatal("local mail did not use the configuration directory")
		}
		letter, err := os.ReadFile(letters[0])
		if err != nil || !strings.Contains(string(letter), "paths@example.test") || !strings.Contains(string(letter), "https://game.example.test/verify") {
			t.Fatal("local mail content does not match configuration")
		}
		if _, err := os.Stat(filepath.Join(cwd, "outgoing letters")); !os.IsNotExist(err) {
			t.Fatal("mail was written relative to caller cwd")
		}
		var count int
		if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email='paths@example.test'`).Scan(&count); err != nil || count != 1 {
			t.Fatalf("registration did not use isolated schema: count=%d err=%v", count, err)
		}
	})
	t.Run("environment only ignores dotenv and uses absolute assets", func(t *testing.T) {
		port, cwd := pathAppPort(t), t.TempDir()
		if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("PRIVATE=\"unclosed\n"), 0600); err != nil {
			t.Fatal(err)
		}
		values := map[string]string{"DATABASE_URL": dsn, "APP_PORT": port, "STATIC_DIR": assets, "MAIL_DIR": filepath.Join(configurationDir, "environment mail")}
		a := startPathApp(t, ctx, binary, cwd, port, values, "-env-file=")
		waitPathApp(t, ctx, a)
		checkPathAppResponse(t, a, "GET", "/", "", 200, "external configuration")
	})
	t.Run("local API starts without dotenv or frontend build", func(t *testing.T) {
		port, cwd := pathAppPort(t), t.TempDir()
		a := startPathApp(t, ctx, binary, cwd, port, map[string]string{"DATABASE_URL": dsn, "APP_PORT": port})
		waitPathApp(t, ctx, a)
		checkPathAppResponse(t, a, "GET", "/healthz", "", 200, "ok")
		checkPathAppResponse(t, a, "GET", "/api/auth/me", "", 401, "unauthenticated")
		checkPathAppResponse(t, a, "GET", "/", "", 404, "404")
	})
	t.Run("production refuses missing frontend with useful safe diagnostics", func(t *testing.T) {
		port := pathAppPort(t)
		a := startPathApp(t, ctx, binary, t.TempDir(), port, map[string]string{"DATABASE_URL": dsn, "APP_PORT": port, "APP_ENV": "production", "PUBLIC_URL": "https://game.example.test", "STATIC_DIR": filepath.Join(t.TempDir(), "private-missing-assets")}, "-env-file=")
		select {
		case <-a.done:
		case <-ctx.Done():
			t.Fatal("production started with missing frontend")
		}
		if a.err == nil {
			t.Fatal("missing production assets accepted")
		}
		output, err := os.ReadFile(a.logFile)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(output), dsn) || strings.Contains(string(output), "private-") {
			t.Fatal("startup diagnostics leaked private values")
		}
		found := false
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			var record map[string]any
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				t.Fatal("startup log is not JSON")
			}
			if record["msg"] == "http_started" {
				t.Fatal("production listened before rejecting missing assets")
			}
			if record["msg"] == "application_failed" && record["stage"] == "routes" && strings.Contains(string(line), "STATIC_DIR") {
				found = true
			}
		}
		if !found {
			t.Fatal("missing actionable STATIC_DIR startup diagnostic")
		}
	})
}
