package handlers

import (
	"context"
	"net/http"

	"gotest/internal/forum"
	"gotest/internal/httpapi"
)

type ForumHandler struct{ service forum.Repository }

func NewForumHandler(service forum.Repository) *ForumHandler { return &ForumHandler{service} }

func (h *ForumHandler) Categories(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Categories(r.Context(), socialActor(r.Context()))
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func requiredForumID(w http.ResponseWriter, r *http.Request, key string, bits int) (int64, bool) {
	id, ok := queryID(w, r, key, bits)
	if !ok {
		return 0, false
	}
	if id == 0 {
		httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Укажите параметр "+key, nil)
		return 0, false
	}
	return id, true
}

func (h *ForumHandler) Topics(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredForumID(w, r, "category_id", 32)
	if !ok {
		return
	}
	in := forum.TopicsInput{CategoryID: int(id)}
	if values, present := r.URL.Query()["cursor"]; present {
		if len(values) != 1 || values[0] == "" {
			httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Некорректный курсор", nil)
			return
		}
		if _, _, err := forum.ParseCursor(values[0]); err != nil {
			httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Некорректный курсор", nil)
			return
		}
		in.Cursor = values[0]
	}
	if values, present := r.URL.Query()["include_hidden"]; present {
		if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			httpapi.WriteError(w, 400, httpapi.CodeInvalidRequest, "Некорректный параметр include_hidden", nil)
			return
		}
		in.IncludeHidden = values[0] == "true"
	}
	result, err := h.service.Topics(r.Context(), socialActor(r.Context()), in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (h *ForumHandler) Topic(w http.ResponseWriter, r *http.Request) {
	id, ok := requiredForumID(w, r, "topic_id", 64)
	if !ok {
		return
	}
	after, ok := queryID(w, r, "after_id", 64)
	if !ok {
		return
	}
	result, err := h.service.Topic(r.Context(), socialActor(r.Context()), id, after)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func forumWrite[Input, Output any](w http.ResponseWriter, r *http.Request, operation func(context.Context, forum.Actor, Input) (Output, error), status int, key string) {
	in, ok := httpapi.DecodeJSONLimit[Input](w, r, 64<<10)
	if !ok {
		return
	}
	result, err := operation(r.Context(), socialActor(r.Context()), *in)
	if err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, map[string]any{key: result})
}

func (h *ForumHandler) Create(w http.ResponseWriter, r *http.Request) {
	forumWrite(w, r, h.service.Create, http.StatusCreated, "topic")
}
func (h *ForumHandler) Reply(w http.ResponseWriter, r *http.Request) {
	forumWrite(w, r, h.service.Reply, http.StatusCreated, "post")
}
func (h *ForumHandler) EditTopic(w http.ResponseWriter, r *http.Request) {
	forumWrite(w, r, h.service.EditTopic, http.StatusOK, "topic")
}
func (h *ForumHandler) EditPost(w http.ResponseWriter, r *http.Request) {
	forumWrite(w, r, h.service.EditPost, http.StatusOK, "post")
}
func (h *ForumHandler) Moderate(w http.ResponseWriter, r *http.Request) {
	in, ok := httpapi.DecodeJSONLimit[forum.ModerateInput](w, r, 64<<10)
	if !ok {
		return
	}
	if err := h.service.Moderate(r.Context(), socialActor(r.Context()), *in); err != nil {
		respondError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"message": "Действие модерации выполнено"})
}
