package routes_test

import (
	"context"
	"encoding/json"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/forum"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
	"gotest/internal/social"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/routers/legacy"
)

func TestForumRoutesRequireAuthentication(t *testing.T) {
	h := router(t, &authStub{}, false, 100)
	for _, tc := range []struct{ method, path string }{
		{"GET", "/api/forum/categories"},
		{"GET", "/api/forum/topics?category_id=1"},
		{"POST", "/api/forum/topics"},
		{"GET", "/api/forum/topic?topic_id=1"},
		{"POST", "/api/forum/replies"},
		{"POST", "/api/forum/edit-topic"},
		{"POST", "/api/forum/edit-reply"},
		{"POST", "/api/forum/moderate"},
	} {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Errorf("%s %s: status=%d, want401", tc.method, tc.path, w.Code)
		}
	}
}

type forumStub struct {
	err    error
	actor  forum.Actor
	calls  int
	filled bool
	input  forum.TopicsInput
}

func (s *forumStub) hit(a forum.Actor) { s.actor = a; s.calls++ }
func (s *forumStub) category() forum.Category {
	return forum.Category{ID: 2, Code: "general", Title: "Общение", Description: "Общий раздел", CreateRole: "user"}
}
func (s *forumStub) topic() forum.Topic {
	now := time.Now().UTC()
	topic := forum.Topic{TopicSummary: forum.TopicSummary{ID: 1, CategoryID: 2, UserID: 7, CharacterID: 3, AuthorName: "Автор", AuthorRole: "user", Title: "Первая тема", CreatedAt: now, UpdatedAt: now}, Body: new("Текст темы"), Version: 1}
	if s.filled {
		topic.PinnedAt = &now
		topic.ClosedAt = &now
		topic.HiddenAt = &now
		topic.EditedAt = &now
		topic.ReplyCount = 2
	}
	return topic
}
func (s *forumStub) post() forum.Post {
	now := time.Now().UTC()
	return forum.Post{ID: 2, TopicID: 1, UserID: 7, CharacterID: 3, AuthorName: "Автор", AuthorRole: "moderator", Body: new("Ответ"), CreatedAt: now, UpdatedAt: now, Version: 1}
}
func (s *forumStub) ban() *social.Ban {
	if !s.filled {
		return nil
	}
	now := time.Now().UTC()
	return &social.Ban{ID: 1, UserID: 7, Scope: "forum", Reason: "Флуд", CreatedByUserID: 8, CreatedAt: now, ExpiresAt: new(now.Add(time.Hour))}
}
func (s *forumStub) Categories(_ context.Context, a forum.Actor) (*forum.CategoryList, error) {
	s.hit(a)
	return &forum.CategoryList{Categories: []forum.Category{s.category()}}, s.err
}
func (s *forumStub) Topics(_ context.Context, a forum.Actor, in forum.TopicsInput) (*forum.TopicList, error) {
	s.hit(a)
	s.input = in
	page := &forum.TopicList{Category: s.category(), Topics: []forum.TopicSummary{}, ForumBan: s.ban()}
	if s.filled {
		page.Topics = append(page.Topics, s.topic().TopicSummary)
		page.NextCursor = new("p:1")
	}
	return page, s.err
}
func (s *forumStub) Topic(_ context.Context, a forum.Actor, _, _ int64) (*forum.TopicPage, error) {
	s.hit(a)
	page := &forum.TopicPage{Topic: s.topic(), Posts: []forum.Post{}, ForumBan: s.ban()}
	if s.filled {
		post := s.post()
		deleted := post
		deleted.ID = 3
		deleted.Body = nil
		deleted.HiddenAt = new(time.Now().UTC())
		deleted.EditedAt = deleted.HiddenAt
		page.Posts = append(page.Posts, post, deleted)
		page.NextAfterID = new(int64(3))
	}
	return page, s.err
}
func (s *forumStub) Create(_ context.Context, a forum.Actor, _ forum.CreateInput) (*forum.Topic, error) {
	s.hit(a)
	topic := s.topic()
	return &topic, s.err
}
func (s *forumStub) Reply(_ context.Context, a forum.Actor, _ forum.ReplyInput) (*forum.Post, error) {
	s.hit(a)
	post := s.post()
	return &post, s.err
}
func (s *forumStub) EditTopic(_ context.Context, a forum.Actor, _ forum.EditTopicInput) (*forum.Topic, error) {
	s.hit(a)
	topic := s.topic()
	return &topic, s.err
}
func (s *forumStub) EditPost(_ context.Context, a forum.Actor, _ forum.EditPostInput) (*forum.Post, error) {
	s.hit(a)
	post := s.post()
	return &post, s.err
}
func (s *forumStub) Moderate(_ context.Context, a forum.Actor, _ forum.ModerateInput) error {
	s.hit(a)
	return s.err
}

func forumRouter(t *testing.T, s *forumStub) http.Handler {
	t.Helper()
	a := &authStub{user: &auth.Identity{ID: 7, RoleID: auth.RoleUser}}
	h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(&characterStub{}), middleware.NewMiddleware(a), routes.Options{
		StaticDir: t.TempDir(), Forum: handlers.NewForumHandler(forum.NewService(s)), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func forumRequest(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.AddCookie(characterCookie())
	r.Header.Set("X-Account-ID", "7")
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestForumTransportRejectsInvalidQueryBeforeService(t *testing.T) {
	for _, path := range []string{
		"/api/forum/topics", "/api/forum/topics?category_id=0", "/api/forum/topics?category_id=1&category_id=2",
		"/api/forum/topics?category_id=2147483648", "/api/forum/topic?topic_id=1&after_id=-1",
		"/api/forum/topic", "/api/forum/topic?topic_id=9223372036854775808",
		"/api/forum/topics?category_id=1&cursor=", "/api/forum/topics?category_id=1&cursor=p:1&cursor=n:2",
		"/api/forum/topics?category_id=1&cursor=n:0", "/api/forum/topics?category_id=1&cursor=p:9223372036854775808",
		"/api/forum/topics?category_id=1&include_hidden=1", "/api/forum/topics?category_id=1&include_hidden=true&include_hidden=false",
	} {
		s := &forumStub{}
		w := httptest.NewRecorder()
		forumRouter(t, s).ServeHTTP(w, forumRequest("GET", path, ""))
		if w.Code != 400 || s.calls != 0 {
			t.Errorf("%s: status%d calls%d", path, w.Code, s.calls)
		}
	}
	s := &forumStub{}
	w := httptest.NewRecorder()
	forumRouter(t, s).ServeHTTP(w, forumRequest("GET", "/api/forum/topics?category_id=2&cursor=n:5&include_hidden=true", ""))
	if w.Code != 200 || s.input != (forum.TopicsInput{CategoryID: 2, Cursor: "n:5", IncludeHidden: true}) {
		t.Fatalf("parsed query=%+v status=%d", s.input, w.Code)
	}
}

func TestForumTransportBindsAccountAndProtectsOrigin(t *testing.T) {
	s := &forumStub{}
	h := forumRouter(t, s)
	r := forumRequest("GET", "/api/forum/categories", "")
	r.Header.Set("X-Account-ID", "8")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 409 || s.calls != 0 {
		t.Fatal("replacement account accepted")
	}
	r = forumRequest("POST", "/api/forum/topics", `{"category_id":2,"character_id":3,"title":"Title","body":"Text"}`)
	r.Header.Set("Origin", "https://other.example.test")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 || s.calls != 0 {
		t.Fatal("cross-origin write accepted")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, forumRequest("GET", "/api/forum/categories", ""))
	if w.Code != 200 || s.actor.UserID != 7 || s.actor.SessionHash == "" {
		t.Fatal("trusted actor not bound to session")
	}
}

func TestForumTransportErrorsAndStrictWrites(t *testing.T) {
	contract, err := legacy.NewRouter(loadOpenAPI(t))
	if err != nil {
		t.Fatal(err)
	}
	valid := `{"topic_id":1,"character_id":3,"body":"Reply"}`
	for _, tc := range []struct {
		body   string
		err    error
		status int
		code   string
	}{
		{valid, nil, 201, ""}, {valid, forum.ErrNotFound, 404, "forum_not_found"},
		{valid, forum.ErrBanned, 403, "forum_banned"}, {valid, forum.ErrClosed, 409, "forum_closed"},
		{valid, forum.ErrConflict, 409, "forum_conflict"}, {valid, forum.ErrTooFast, 429, "rate_limited"},
		{valid, social.ErrCharacterChanged, 409, "character_changed"}, {valid, social.ErrForbidden, 403, "forbidden"},
		{valid, auth.ErrUnauthenticated, 401, "unauthenticated"}, {valid, errors.New("private backend detail"), 500, "internal_error"},
		{`{"topic_id":1,"character_id":3,"body":"Reply","user_id":9}`, nil, 400, "invalid_json"},
		{`{"topic_id":1,"body":"Reply"}`, nil, 422, "validation_failed"},
	} {
		s := &forumStub{err: tc.err}
		r := forumRequest("POST", "/api/forum/replies", tc.body)
		w := httptest.NewRecorder()
		forumRouter(t, s).ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("status=%d want%d body=%s", w.Code, tc.status, w.Body.String())
		}
		if tc.code != "" && !strings.Contains(w.Body.String(), `"code":"`+tc.code+`"`) {
			t.Fatalf("wrong error: %s", w.Body.String())
		}
		if strings.Contains(w.Body.String(), "private backend detail") {
			t.Fatal("internal detail exposed")
		}
		if tc.status == 429 && w.Header().Get("Retry-After") != "5" {
			t.Fatal("wrong cooldown")
		}
		if err := validateContractResponse(contractInput(t, contract, r), w.Code, w.Header(), w.Body.Bytes()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestForumTransportAllowsUnicodeLimitAndBoundsRequest(t *testing.T) {
	for _, tc := range []struct{ count, status int }{{10000, 201}, {10001, 422}, {20000, 413}} {
		payload, err := json.Marshal(forum.ReplyInput{TopicID: 1, CharacterID: 3, Body: strings.Repeat("😀", tc.count)})
		if err != nil {
			t.Fatal(err)
		}
		s := &forumStub{}
		w := httptest.NewRecorder()
		forumRouter(t, s).ServeHTTP(w, forumRequest("POST", "/api/forum/replies", string(payload)))
		if w.Code != tc.status {
			t.Errorf("runes%d: status%d want%d", tc.count, w.Code, tc.status)
		}
	}
}
