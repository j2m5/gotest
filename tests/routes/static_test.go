package routes_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
)

func staticRouter(directory string, required bool) (http.Handler, error) {
	s := &authStub{}
	return routes.New(handlers.NewHandler(s, false, s), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(s), routes.Options{StaticDir: directory, RequireStatic: required})
}

func TestRequiredStaticFrontendRejectsIncompleteDeployment(t *testing.T) {
	for _, kind := range []string{"missing directory", "missing index", "directory index", "empty index"} {
		t.Run(kind, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "private-assets")
			if kind != "missing directory" {
				if err := os.Mkdir(directory, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "directory index" {
				if err := os.Mkdir(filepath.Join(directory, "index.html"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "empty index" {
				if err := os.WriteFile(filepath.Join(directory, "index.html"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := staticRouter(directory, true); err == nil || !strings.Contains(err.Error(), "STATIC_DIR") || strings.Contains(err.Error(), "private-") {
				t.Fatalf("invalid production assets = %v", err)
			}
		})
	}
}

func TestOptionalStaticFrontendKeepsDevelopmentAPIAvailable(t *testing.T) {
	h, err := staticRouter(filepath.Join(t.TempDir(), "unbuilt"), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		path   string
		status int
		body   string
	}{
		{"/healthz", 200, "ok"}, {"/api/missing", 404, "route_not_found"}, {"/", 404, "404"},
	} {
		w := request(h, "GET", tt.path, "", nil)
		if w.Code != tt.status || !strings.Contains(w.Body.String(), tt.body) {
			t.Fatalf("%s = %d %s", tt.path, w.Code, w.Body)
		}
	}
}

func TestStaticRootIsFixedBeforeWorkingDirectoryChanges(t *testing.T) {
	base := t.TempDir()
	directory := filepath.Join(base, "web assets")
	if err := os.MkdirAll(filepath.Join(directory, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"index.html": "<html>fixed root</html>", "assets/client.js": "fixed_asset"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(base)
	h, err := staticRouter("web assets", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	for _, tt := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", "/", 200, "fixed root"}, {"GET", "/characters/new", 200, "fixed root"},
		{"GET", "/assets/client.js", 200, "fixed_asset"}, {"HEAD", "/assets/client.js", 200, ""},
		{"GET", "/assets/missing.js", 404, "404"}, {"GET", "/api/missing", 404, "route_not_found"},
		{"POST", "/characters/new", 405, "Method not allowed"},
	} {
		w := request(h, tt.method, tt.path, "", nil)
		if w.Code != tt.status || !strings.Contains(w.Body.String(), tt.body) {
			t.Fatalf("%s %s = %d %s", tt.method, tt.path, w.Code, w.Body)
		}
		if tt.method == "HEAD" && w.Body.Len() != 0 {
			t.Fatal("HEAD returned a response body")
		}
	}
}
