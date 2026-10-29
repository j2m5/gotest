package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"sort"

	"gotest/internal/validation"
)

const (
	CodeInvalidJSON          = "invalid_json"
	CodeInvalidRequest       = "invalid_request"
	CodeUnsupportedMediaType = "unsupported_media_type"
	CodeRequestTooLarge      = "request_too_large"
	CodeMethodNotAllowed     = "method_not_allowed"
	CodeRouteNotFound        = "route_not_found"
	CodeOriginForbidden      = "origin_forbidden"
	CodeRateLimited          = "rate_limited"
	CodeInternalError        = "internal_error"
	CodeUnauthenticated      = "unauthenticated"
	CodeAccountChanged       = "account_changed"
	CodeValidationFailed     = "validation_failed"
)

type errorResponse struct {
	Error  string            `json:"error"`
	Code   string            `json:"code"`
	Fields validation.Fields `json:"fields,omitempty"`
	Errors []string          `json:"errors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		writeErrorBody(w, http.StatusInternalServerError, CodeInternalError, "Внутренняя ошибка сервера", nil)
		return
	}
	body = append(body, '\n')
	setJSONHeaders(w)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func WriteError(w http.ResponseWriter, status int, code, message string, fields validation.Fields) {
	writeErrorBody(w, status, code, message, fields)
}

func writeErrorBody(w http.ResponseWriter, status int, code, message string, fields validation.Fields) {
	response := errorResponse{Error: message, Code: code, Fields: fields}
	for _, key := range sortedFieldKeys(fields) {
		response.Errors = append(response.Errors, fields[key]...)
	}
	WriteJSON(w, status, response)
}

func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
}

func sortedFieldKeys(fields validation.Fields) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	return DecodeJSONLimit[T](w, r, 16<<10)
}

// DecodeJSONLimit keeps strict decoding while allowing a route to choose its
// byte budget. The limit is application-owned, never taken from client input.
func DecodeJSONLimit[T any](w http.ResponseWriter, r *http.Request, limit int64) (*T, bool) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		WriteError(w, http.StatusUnsupportedMediaType, CodeUnsupportedMediaType, "Ожидается Content-Type: application/json", nil)
		return nil, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	var raw json.RawMessage
	if err = decoder.Decode(&raw); err != nil {
		writeDecodeError(w, err)
		return nil, false
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		writeDecodeError(w, errors.New("expected JSON object"))
		return nil, false
	}
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	var input *T
	if err = strict.Decode(&input); err != nil {
		writeDecodeError(w, err)
		return nil, false
	}
	if input == nil {
		writeDecodeError(w, errors.New("expected JSON object"))
		return nil, false
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = errors.New("extra JSON value")
		}
		writeDecodeError(w, err)
		return nil, false
	}
	return input, true
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, CodeRequestTooLarge, "Тело запроса слишком большое", nil)
		return
	}
	WriteError(w, http.StatusBadRequest, CodeInvalidJSON, "Некорректный JSON-запрос", nil)
}
