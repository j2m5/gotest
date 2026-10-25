package httpapi_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest/internal/httpapi"
	"gotest/internal/validation"
)

type sample struct {
	Name string `json:"name"`
}

func TestDecodeJSONRejectsMalformedVariants(t *testing.T) {
	cases := []struct{ name, body string }{
		{"malformed", `{"name":`}, {"null", `null`}, {"unknown field", `{"extra":true}`},
		{"second value", `{} {}`}, {"trailing whitespace after large body", `{}` + strings.Repeat(" ", 16<<10)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json; charset=utf-8")
			w := httptest.NewRecorder()
			got, ok := httpapi.DecodeJSON[sample](w, r)
			if ok || got != nil {
				t.Fatalf("got %#v, %v; want nil,false", got, ok)
			}
			if w.Code != 400 && w.Code != 413 {
				t.Fatalf("status %d", w.Code)
			}
			var env map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if env["code"] != httpapi.CodeInvalidJSON && env["code"] != httpapi.CodeRequestTooLarge {
				t.Fatalf("code %v", env["code"])
			}
		})
	}
}

func TestDecodeJSONMediaTypeAndLimit(t *testing.T) {
	for _, tc := range []struct {
		name, contentType, body, code string
		status                        int
	}{
		{"wrong type", "text/plain", `{}`, httpapi.CodeUnsupportedMediaType, 415},
		{"too large", "application/json", `{"name":"` + strings.Repeat("x", 16<<10) + `"}`, httpapi.CodeRequestTooLarge, 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			_, ok := httpapi.DecodeJSON[sample](w, r)
			if ok || w.Code != tc.status {
				t.Fatalf("ok=%v status=%d", ok, w.Code)
			}
			var env map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &env)
			if env["code"] != tc.code {
				t.Fatalf("code=%v", env["code"])
			}
		})
	}
}

func TestWriteErrorFlattensFieldsInSortedOrder(t *testing.T) {
	w := httptest.NewRecorder()
	httpapi.WriteError(w, 422, httpapi.CodeValidationFailed, "Check fields", validation.Fields{"z": {"z1", "z2"}, "a": {"a1"}})
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("headers: %v", w.Header())
	}
	var got struct {
		Error  string            `json:"error"`
		Code   string            `json:"code"`
		Fields validation.Fields `json:"fields"`
		Errors []string          `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != httpapi.CodeValidationFailed || strings.Join(got.Errors, ",") != "a1,z1,z2" {
		t.Fatalf("body: %+v", got)
	}
}

func TestWriteJSONEncodingFailureIsGenericAndNoPartialSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	httpapi.WriteJSON(w, 200, map[string]any{"secret": "hidden", "bad": make(chan int)})
	if w.Code != 500 {
		t.Fatalf("status %d", w.Code)
	}
	var env map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env["code"] != httpapi.CodeInternalError || strings.Contains(w.Body.String(), "hidden") {
		t.Fatalf("body %s", w.Body.String())
	}
}
