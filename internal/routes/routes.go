package routes

import (
	"context"
	"gotest/internal/handlers"
	"gotest/internal/httpapi"
	"gotest/internal/middleware"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	StaticDir, TrustedOrigin string
	AuthLimit                int
}

func New(h *handlers.Handler, characters *handlers.CharacterHandler, m *middleware.Middleware, options Options) (http.Handler, error) {
	api := http.NewServeMux()
	limiter := middleware.NewRateLimiter(options.AuthLimit, time.Minute)
	bind := func(route, method string, handler http.HandlerFunc, limited bool) {
		endpoint := methodHandler(map[string]http.HandlerFunc{method: handler})
		if limited {
			endpoint = limiter.Handler(endpoint)
		}
		api.Handle(route, endpoint)
	}
	bind("/api/auth/register", "POST", h.Register, true)
	bind("/api/auth/login", "POST", h.Login, true)
	bind("/api/auth/verify", "POST", h.VerifyEmail, true)
	bind("/api/auth/resend-verification", "POST", h.ResendVerification, true)
	bind("/api/auth/logout", "POST", h.Logout, false)
	bind("/api/auth/me", "GET", m.Auth(h.Me), false)
	api.Handle("/api/characters", methodHandler(map[string]http.HandlerFunc{
		"GET":  m.Auth(characters.List),
		"POST": m.Auth(characters.Create),
	}))
	bind("/api/characters/catalog", "GET", m.Auth(characters.Catalog), false)
	bind("/api/characters/active", "PUT", m.Auth(characters.SetActive), false)
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, 404, httpapi.CodeRouteNotFound, "API-маршрут не найден", nil)
	})
	mux := http.NewServeMux()
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		api.ServeHTTP(w, r.WithContext(ctx))
	})
	mux.Handle("/api/", apiHandler)
	mux.Handle("/api", apiHandler)
	if options.StaticDir == "" {
		options.StaticDir = "frontend/dist"
	}
	fileServer := http.FileServer(http.Dir(options.StaticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", 405)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		info, err := os.Stat(filepath.Join(options.StaticDir, filepath.FromSlash(name)))
		if err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(name, "assets/") || filepath.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(options.StaticDir, "index.html"))
	})
	protection := http.NewCrossOriginProtection()
	if options.TrustedOrigin != "" {
		if err := protection.AddTrustedOrigin(options.TrustedOrigin); err != nil {
			return nil, err
		}
	}
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, 403, httpapi.CodeOriginForbidden, "Запрос с другого сайта запрещён", nil)
	}))
	protected := protection.Handler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		protected.ServeHTTP(w, r)
	}), nil
}

func methodHandler(endpoints map[string]http.HandlerFunc) http.Handler {
	methods := make([]string, 0, len(endpoints))
	for method := range endpoints {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	allow := strings.Join(methods, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint, ok := endpoints[r.Method]
		if !ok {
			w.Header().Set("Allow", allow)
			httpapi.WriteError(w, 405, httpapi.CodeMethodNotAllowed, "Метод не поддерживается", nil)
			return
		}
		endpoint(w, r)
	})
}
