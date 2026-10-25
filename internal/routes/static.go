package routes

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func newStaticHandler(directory string, required bool) (http.Handler, error) {
	if directory == "" {
		directory = "frontend/dist"
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return nil, errors.New("STATIC_DIR could not be resolved")
	}
	index := filepath.Join(root, "index.html")
	if required {
		if err := validateIndex(index); err != nil {
			return nil, err
		}
	}
	fileServer := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", 405)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(name)))
		if err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(name, "assets/") || filepath.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, index)
	}), nil
}

func validateIndex(filename string) error {
	invalid := errors.New("STATIC_DIR must contain a readable, nonempty index.html file in production")
	info, err := os.Stat(filename)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return invalid
	}
	file, err := os.Open(filename)
	if err != nil {
		return invalid
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return invalid
	}
	return nil
}
