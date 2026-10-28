package routes_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	openapirouters "github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/realtime"
	"gotest/internal/routes"
	"gotest/internal/validation"
)

func loadOpenAPI(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestOpenAPIDocument(t *testing.T) {
	doc := loadOpenAPI(t)
	ids := map[string]bool{}
	for path, item := range doc.Paths.Map() {
		for method, operation := range item.Operations() {
			if operation.OperationID == "" || ids[operation.OperationID] {
				t.Fatalf("%s %s: missing or duplicate operationId", method, path)
			}
			ids[operation.OperationID] = true
		}
	}
	if len(ids) == 0 {
		t.Fatal("the API document has no operations")
	}
}

// Reuse the route suite's service doubles, with populated collections so that
// array item schemas and both null/non-null DTO variants are actually exercised.
type contractCharacters struct {
	characterStub
	empty, populated bool
}

func (s *contractCharacters) sample() *character.Character {
	c := &character.Character{ID: 10, Name: "Malgus", FactionID: 2, ClassID: 8,
		Role: "damage", Level: 1, Experience: 0, HP: 100, HPCurrent: 100, MP: 50, MPCurrent: 50}
	if s.populated {
		c.Avatar = new("/avatars/sith.png")
	}
	return c
}

func (s *contractCharacters) List(context.Context, character.Actor) (*character.Roster, error) {
	roster := &character.Roster{Characters: []character.Character{}}
	if !s.empty {
		roster.Characters = append(roster.Characters, *s.sample())
		roster.ActiveCharacterID = new(10)
	}
	return roster, s.err
}

func (s *contractCharacters) Catalog(ctx context.Context) (*character.Catalog, error) {
	catalog, err := s.characterStub.Catalog(ctx)
	if s.populated {
		catalog.Factions[0].Description, catalog.Factions[0].Icon = new("Empire"), new("empire.png")
		catalog.Classes[0].Description, catalog.Classes[0].Icon = new("Sith"), new("sith.png")
	}
	return catalog, err
}

func (s *contractCharacters) Create(context.Context, character.Actor, character.CreateInput) (*character.Character, error) {
	return s.sample(), s.err
}

func (s *contractCharacters) SetActive(_ context.Context, _ character.Actor, id *int) (*character.Character, error) {
	if id == nil {
		return nil, s.err
	}
	return s.sample(), s.err
}

type contractCase struct {
	name, method, path, body string
	status                   int
	authErr, characterErr    error
	noCookie, empty, filled  bool
	invalidRequest           bool
	headers                  http.Header
	notReady, limited        bool
}

func contractServer(t *testing.T, tc contractCase) *httptest.Server {
	t.Helper()
	identity := &auth.Identity{ID: 1, Email: "p@example.test", Login: "player"}
	if tc.filled {
		identity.CharacterID = new(10)
	}
	a := &authStub{user: identity, err: tc.authErr}
	lifecycle, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	hub := realtime.New(lifecycle, a, "http://localhost:5173")
	c := &contractCharacters{characterStub: characterStub{err: tc.characterErr}, empty: tc.empty, populated: tc.filled}
	limit := 1000
	if tc.limited {
		limit = 1
	}
	h, err := routes.New(handlers.NewHandler(a, true, a), handlers.NewCharacterHandler(c), middleware.NewMiddleware(a), routes.Options{
		StaticDir: t.TempDir(), AuthLimit: limit, TrustedOrigin: "http://localhost:5173",
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Social: handlers.NewSocialHandler(&socialStub{filled: tc.filled}, hub.Notify), ChatEvents: hub.ServeHTTP, Lifecycle: lifecycle,
		Ready: func(context.Context) error {
			if tc.notReady {
				return errors.New("database offline")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(h)
	s.Client().Timeout = 5 * time.Second
	t.Cleanup(s.Close)
	return s
}

func contractRequest(t *testing.T, server *httptest.Server, tc contractCase) *http.Request {
	t.Helper()
	r, err := http.NewRequest(tc.method, server.URL+tc.path, strings.NewReader(tc.body))
	if err != nil {
		t.Fatal(err)
	}
	if tc.body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if !tc.noCookie {
		r.AddCookie(characterCookie())
	}
	for name, values := range tc.headers {
		r.Header.Del(name)
		for _, value := range values {
			r.Header.Add(name, value)
		}
	}
	return r
}

func contractOptions() *openapi3filter.Options {
	return &openapi3filter.Options{
		IncludeResponseStatus: true, RejectWhenRequestBodyNotSpecified: true,
		// Authentication itself is exercised by the real middleware. Here enforce
		// that cases using a secured operation supply its documented credential.
		AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
			cookie, err := input.RequestValidationInput.Request.Cookie(input.SecurityScheme.Name)
			if err != nil || cookie.Value == "" {
				return errors.New("missing session cookie")
			}
			return nil
		},
	}
}

func contractInput(t *testing.T, router openapirouters.Router, r *http.Request) *openapi3filter.RequestValidationInput {
	t.Helper()
	// The document uses a same-origin server URL ("/"). The legacy matcher
	// expects an origin-relative request URL for that server, not httptest's host.
	match := r.Clone(r.Context())
	match.URL.Scheme, match.URL.Host = "", ""
	route, params, err := router.FindRoute(match)
	if err != nil {
		t.Fatal(err)
	}
	return &openapi3filter.RequestValidationInput{Request: r, Route: route, PathParams: params, Options: contractOptions()}
}

func validateContractResponse(input *openapi3filter.RequestValidationInput, status int, header http.Header, body []byte) error {
	response := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input, Status: status, Header: header, Options: contractOptions(),
	}
	response.SetBodyBytes(body)
	return openapi3filter.ValidateResponse(context.Background(), response)
}

func contractExchange(t *testing.T, server *httptest.Server, r *http.Request, status int) (*http.Response, []byte) {
	t.Helper()
	response, err := server.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("status %d, want %d: %s", response.StatusCode, status, body)
	}
	if r.Method == http.MethodHead && len(body) != 0 {
		t.Fatal("HEAD response must not have a body")
	}
	return response, body
}

func TestOpenAPIOperations(t *testing.T) {
	doc := loadOpenAPI(t)
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("a", 64)
	cases := []contractCase{
		{name: "register", method: "POST", path: "/api/auth/register", body: `{"email":"p@example.test","login":"player","password":"password"}`, status: 201},
		{name: "login session", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"password"}`, status: 200},
		{name: "login remembered", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"password","remember_me":true}`, status: 200},
		{name: "login null remember", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"password","remember_me":null}`, status: 200},
		{name: "logout", method: "POST", path: "/api/auth/logout", status: 200},
		{name: "logout no cookie", method: "POST", path: "/api/auth/logout", noCookie: true, status: 200},
		{name: "current account", method: "GET", path: "/api/auth/me", status: 200},
		{name: "current account selected", method: "GET", path: "/api/auth/me", filled: true, status: 200},
		{name: "verify JSON", method: "POST", path: "/api/auth/verify", body: `{"token":"` + token + `"}`, status: 200},
		{name: "verify query", method: "POST", path: "/api/auth/verify?token=" + token, status: 200},
		{name: "resend verification", method: "POST", path: "/api/auth/resend-verification", body: `{"email":"p@example.test","password":"password"}`, status: 200},
		{name: "forgot password", method: "POST", path: "/api/auth/forgot-password", body: `{"email":"p@example.test"}`, status: 200},
		{name: "reset password", method: "POST", path: "/api/auth/reset-password", body: `{"token":"` + token + `","password":"password"}`, status: 200},
		{name: "empty roster", method: "GET", path: "/api/characters", empty: true, status: 200},
		{name: "roster null avatar", method: "GET", path: "/api/characters", status: 200},
		{name: "roster with avatar", method: "GET", path: "/api/characters", filled: true, status: 200},
		{name: "catalog null fields", method: "GET", path: "/api/characters/catalog", status: 200},
		{name: "catalog filled fields", method: "GET", path: "/api/characters/catalog", filled: true, status: 200},
		{name: "create character", method: "POST", path: "/api/characters", body: `{"name":"Malgus","faction_id":2,"class_id":8}`, status: 201},
		{name: "select character", method: "PUT", path: "/api/characters/active", body: `{"character_id":10}`, status: 200},
		{name: "clear character", method: "PUT", path: "/api/characters/active", body: `{"character_id":null}`, status: 200},
		{name: "chat empty", method: "GET", path: "/api/chat/messages", status: 200},
		{name: "moderators empty", method: "GET", path: "/api/chat/moderators", status: 200},
		{name: "moderators selected and unselected characters", method: "GET", path: "/api/chat/moderators", filled: true, status: 200},
		{name: "chat messages and ban", method: "GET", path: "/api/chat/messages?before_id=50", filled: true, status: 200},
		{name: "chat send", method: "POST", path: "/api/chat/messages", body: `{"body":"Hello","character_id":3}`, status: 201},
		{name: "chat delete", method: "POST", path: "/api/chat/delete", body: `{"message_id":1,"reason":"Spam"}`, status: 200},
		{name: "bans", method: "GET", path: "/api/moderation/bans?user_id=2", filled: true, status: 200},
		{name: "ban", method: "POST", path: "/api/moderation/bans", body: `{"user_id":2,"scope":"account","duration_hours":null,"reason":"Spam"}`, status: 201},
		{name: "revoke ban", method: "POST", path: "/api/moderation/revoke", body: `{"ban_id":1,"reason":"Appeal"}`, status: 200},
		{name: "moderation journal", method: "GET", path: "/api/moderation/actions", filled: true, status: 200},
		{name: "chat events", method: "GET", path: "/api/chat/events?account_id=1", status: 101},
		{name: "health", method: "GET", path: "/healthz", status: 200},
		{name: "health HEAD", method: "HEAD", path: "/healthz", status: 200},
		{name: "ready", method: "GET", path: "/readyz", status: 200},
		{name: "ready HEAD", method: "HEAD", path: "/readyz", status: 200},
	}
	covered := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := contractServer(t, tc)
			r := contractRequest(t, server, tc)
			input := contractInput(t, router, r)
			if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
				t.Fatalf("request: %v", err)
			}
			var response *http.Response
			var body []byte
			if tc.status == 101 {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				r.Header.Set("Origin", server.URL)
				c, upgrade, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(r.URL.String(), "http"), &websocket.DialOptions{HTTPHeader: r.Header})
				if err != nil {
					t.Fatal(err)
				}
				defer c.CloseNow()
				_, data, err := c.Read(ctx)
				if err != nil || string(data) != `{"type":"changed"}` {
					t.Fatalf("socket event: %s %v", data, err)
				}
				response = upgrade
			} else {
				response, body = contractExchange(t, server, r, tc.status)
			}
			if err := validateContractResponse(input, response.StatusCode, response.Header, body); err != nil {
				t.Fatalf("response: %v", err)
			}
			covered[input.Route.Operation.OperationID] = true
			if tc.path == "/api/auth/login" || tc.path == "/api/auth/logout" || tc.path == "/api/auth/reset-password" {
				cookies := response.Cookies()
				if len(cookies) != 1 || cookies[0].Name != middleware.SessionCookie || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].Path != "/" || cookies[0].SameSite != http.SameSiteLaxMode {
					t.Fatal("invalid session cookie attributes")
				}
				cookie := cookies[0]
				switch {
				case tc.name == "login remembered":
					if cookie.MaxAge <= 0 || cookie.Expires.IsZero() {
						t.Fatal("missing persistent expiry")
					}
				case tc.path == "/api/auth/login":
					if cookie.MaxAge != 0 || !cookie.Expires.IsZero() {
						t.Fatal("ordinary cookie must not persist")
					}
				default:
					if cookie.Value != "" || cookie.MaxAge != -1 {
						t.Fatal("cookie must be expired")
					}
				}
			}
		})
	}
	for path, item := range doc.Paths.Map() {
		for method, operation := range item.Operations() {
			if !covered[operation.OperationID] {
				t.Errorf("%s %s needs a successful HTTP contract scenario", method, path)
			}
		}
	}
}

func TestOpenAPIErrorResponses(t *testing.T) {
	doc := loadOpenAPI(t)
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	credentials := `{"email":"p@example.test","password":"password"}`
	registration := `{"email":"p@example.test","login":"player","password":"password"}`
	token := strings.Repeat("a", 64)
	reset := `{"token":"` + token + `","password":"password"}`
	create := `{"name":"Malgus","faction_id":2,"class_id":8}`
	invalid := &validation.Error{Fields: validation.Fields{"email": {"Invalid email"}}}
	failure := errors.New("internal detail")
	cases := []contractCase{
		{name: "register conflict", method: "POST", path: "/api/auth/register", body: registration, authErr: auth.ErrConflict, status: 409},
		{name: "register invalid", method: "POST", path: "/api/auth/register", body: registration, authErr: invalid, status: 422},
		{name: "register internal", method: "POST", path: "/api/auth/register", body: registration, authErr: failure, status: 500},
		{name: "wrong credentials", method: "POST", path: "/api/auth/login", body: credentials, authErr: auth.ErrInvalidCredentials, status: 401},
		{name: "unverified", method: "POST", path: "/api/auth/login", body: credentials, authErr: auth.ErrUnverified, status: 403},
		{name: "login internal", method: "POST", path: "/api/auth/login", body: credentials, authErr: failure, status: 500},
		{name: "logout internal", method: "POST", path: "/api/auth/logout", authErr: failure, status: 500},
		{name: "logout account changed", method: "POST", path: "/api/auth/logout", headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "logout invalid account", method: "POST", path: "/api/auth/logout", headers: http.Header{"X-Account-ID": {"bad"}}, invalidRequest: true, status: 400},
		{name: "me account changed", method: "GET", path: "/api/auth/me", headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "roster account changed", method: "GET", path: "/api/characters", headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "catalog account changed", method: "GET", path: "/api/characters/catalog", headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "create account changed", method: "POST", path: "/api/characters", body: create, headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "selection account changed", method: "PUT", path: "/api/characters/active", body: `{"character_id":null}`, headers: http.Header{"X-Account-ID": {"2"}}, status: 409},
		{name: "me no cookie", method: "GET", path: "/api/auth/me", noCookie: true, invalidRequest: true, status: 401},
		{name: "me internal", method: "GET", path: "/api/auth/me", authErr: failure, status: 500},
		{name: "bad verification token", method: "POST", path: "/api/auth/verify", body: `{"token":"` + token + `"}`, authErr: auth.ErrInvalidToken, status: 400},
		{name: "resend bad credentials", method: "POST", path: "/api/auth/resend-verification", body: credentials, authErr: auth.ErrInvalidCredentials, status: 401},
		{name: "already verified", method: "POST", path: "/api/auth/resend-verification", body: credentials, authErr: auth.ErrAlreadyVerified, status: 409},
		{name: "outbox unavailable", method: "POST", path: "/api/auth/resend-verification", body: credentials, authErr: failure, status: 500},
		{name: "forgot invalid email", method: "POST", path: "/api/auth/forgot-password", body: `{"email":"bad"}`, authErr: invalid, status: 422},
		{name: "reset queue unavailable", method: "POST", path: "/api/auth/forgot-password", body: `{"email":"p@example.test"}`, authErr: failure, status: 503},
		{name: "bad reset token", method: "POST", path: "/api/auth/reset-password", body: reset, authErr: auth.ErrInvalidResetToken, status: 400},
		{name: "reset invalid password", method: "POST", path: "/api/auth/reset-password", body: reset, authErr: &validation.Error{Fields: validation.Fields{"password": {"Too short"}}}, status: 422},
		{name: "reset internal", method: "POST", path: "/api/auth/reset-password", body: reset, authErr: failure, status: 500},
		{name: "character missing", method: "PUT", path: "/api/characters/active", body: `{"character_id":10}`, characterErr: character.ErrNotFound, status: 404},
		{name: "name taken", method: "POST", path: "/api/characters", body: create, characterErr: character.ErrNameTaken, status: 409},
		{name: "class mismatch", method: "POST", path: "/api/characters", body: create, characterErr: character.ErrInvalidClass, status: 422},
		{name: "character validation", method: "POST", path: "/api/characters", body: create, characterErr: &validation.Error{Fields: validation.Fields{"name": {"Too short"}}}, status: 422},
		{name: "session expired in repository", method: "GET", path: "/api/characters", characterErr: character.ErrUnauthenticated, status: 401},
		{name: "character internal", method: "GET", path: "/api/characters", characterErr: failure, status: 500},
		{name: "catalog internal", method: "GET", path: "/api/characters/catalog", characterErr: failure, status: 500},
		{name: "invalid JSON", method: "POST", path: "/api/auth/login", body: `{`, invalidRequest: true, status: 400},
		{name: "unknown fields", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"password","role_id":1}`, invalidRequest: true, status: 400},
		{name: "wrong field type", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"password","remember_me":"yes"}`, invalidRequest: true, status: 400},
		{name: "missing selection", method: "PUT", path: "/api/characters/active", body: `{}`, invalidRequest: true, status: 400},
		{name: "wrong content type", method: "POST", path: "/api/auth/login", body: credentials, headers: http.Header{"Content-Type": {"text/plain"}}, invalidRequest: true, status: 415},
		{name: "body too large", method: "POST", path: "/api/auth/login", body: `{"email":"p@example.test","password":"` + strings.Repeat("a", 17<<10) + `"}`, status: 413},
		{name: "origin rejected", method: "POST", path: "/api/auth/login", body: credentials, headers: http.Header{"Sec-Fetch-Site": {"cross-site"}}, status: 403},
		{name: "rate limited", method: "POST", path: "/api/auth/login", body: credentials, limited: true, status: 429},
		{name: "not ready", method: "GET", path: "/readyz", notReady: true, status: 503},
		{name: "not ready HEAD", method: "HEAD", path: "/readyz", notReady: true, status: 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := contractServer(t, tc)
			r := contractRequest(t, server, tc)
			input := contractInput(t, router, r)
			err := openapi3filter.ValidateRequest(context.Background(), input)
			if (err != nil) != tc.invalidRequest {
				t.Fatalf("request validity differs from scenario (invalid=%v): %v", tc.invalidRequest, err)
			}
			// Rebuild the request: a failed decoder may have consumed its body.
			r = contractRequest(t, server, tc)
			if tc.limited {
				contractExchange(t, server, contractRequest(t, server, tc), 200)
			}
			response, body := contractExchange(t, server, r, tc.status)
			if err := validateContractResponse(input, response.StatusCode, response.Header, body); err != nil {
				t.Fatalf("response: %v", err)
			}
		})
	}
}

func TestOpenAPIRoutingErrors(t *testing.T) {
	doc := loadOpenAPI(t)
	server := contractServer(t, contractCase{})
	for path, item := range doc.Paths.Map() {
		t.Run(path, func(t *testing.T) {
			r := contractRequest(t, server, contractCase{method: "DELETE", path: path})
			response, body := contractExchange(t, server, r, 405)
			responses := openapi3.NewResponses()
			responses.Set("405", doc.Components.Responses["MethodNotAllowed"])
			input := &openapi3filter.RequestValidationInput{Request: r, Route: &openapirouters.Route{
				Spec: doc, Operation: &openapi3.Operation{Responses: responses},
			}}
			if err := validateContractResponse(input, response.StatusCode, response.Header, body); err != nil {
				t.Fatal(err)
			}
			methods := []string{}
			for method := range item.Operations() {
				methods = append(methods, method)
			}
			sort.Strings(methods)
			if allow := response.Header.Get("Allow"); allow != strings.Join(methods, ", ") {
				t.Errorf("Allow=%q differs from documented methods %v", allow, methods)
			}
		})
	}
	r := contractRequest(t, server, contractCase{method: "GET", path: "/api/not-a-route"})
	response, body := contractExchange(t, server, r, 404)
	responses := openapi3.NewResponses()
	responses.Set("404", doc.Components.Responses["NotFound"])
	input := &openapi3filter.RequestValidationInput{Request: r, Route: &openapirouters.Route{
		Spec: doc, Operation: &openapi3.Operation{Responses: responses},
	}}
	if err := validateContractResponse(input, response.StatusCode, response.Header, body); err != nil {
		t.Fatal(err)
	}
}

func TestOpenAPIRejectsResponseDrift(t *testing.T) {
	doc := loadOpenAPI(t)
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	tc := contractCase{method: "GET", path: "/api/auth/me"}
	server := contractServer(t, tc)
	r := contractRequest(t, server, tc)
	input := contractInput(t, router, r)
	response, body := contractExchange(t, server, r, 200)
	if err := validateContractResponse(input, 200, response.Header, body); err != nil {
		t.Fatal(err)
	}
	// Start from an actual HTTP response, then introduce typical accidental DTO
	// regressions. If the harness ever skips validation these probes must fail.
	for _, change := range []string{"missing nullable field", "wrong type", "private field", "unexpected status", "missing header", "wrong content type"} {
		t.Run(change, func(t *testing.T) {
			var decoded map[string]any
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatal(err)
			}
			user := decoded["user"].(map[string]any)
			header, status := response.Header.Clone(), 200
			switch change {
			case "missing nullable field":
				delete(user, "character_id")
			case "wrong type":
				user["id"] = "1"
			case "private field":
				user["password_hash"] = "must-not-be-public"
			case "unexpected status":
				status = 202
			case "missing header":
				header.Del("Cache-Control")
			case "wrong content type":
				header.Set("Content-Type", "text/plain")
			default:
				t.Fatal(fmt.Sprintf("unknown mutation %q", change))
			}
			changed, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateContractResponse(input, status, header, changed); err == nil {
				t.Fatal("incompatible response was accepted")
			}
		})
	}
}
