package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	tinyserp "github.com/okayama-daiki/tiny-serp"
)

func TestHandlerSearchSuccess(t *testing.T) {
	html := readFixture(t, "../testdata/bing.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=bing&q=aws+lambda", nil)

	NewHandler(client, tinyserp.DefaultEngines()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", contentType)
	}

	var response tinyserp.SearchResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.SearchInformation.Engine != "bing" {
		t.Fatalf("unexpected engine: %s", response.SearchInformation.Engine)
	}
	if len(response.Items) != 2 {
		t.Fatalf("unexpected items length: %d", len(response.Items))
	}
}

func TestHandlerSearchValidation(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})}

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantError  string
	}{
		{
			name:       "rejects non get requests",
			method:     http.MethodPost,
			target:     "/search?engine=bing&q=aws+lambda",
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method must be GET",
		},
		{
			name:       "requires query",
			method:     http.MethodGet,
			target:     "/search?engine=bing",
			wantStatus: http.StatusBadRequest,
			wantError:  "query parameter q is required",
		},
		{
			name:       "requires supported engine",
			method:     http.MethodGet,
			target:     "/search?engine=google&q=aws+lambda",
			wantStatus: http.StatusBadRequest,
			wantError:  "unsupported engine: google",
		},
	}

	handler := NewHandler(client, tinyserp.DefaultEngines())
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.target, nil)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("unexpected status code: %d", recorder.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if response["error"] != test.wantError {
				t.Fatalf("unexpected error message: %s", response["error"])
			}
		})
	}
}

func TestHandlerMapsBlockedUpstreamToBadGateway(t *testing.T) {
	html := readFixture(t, "../testdata/duckduckgo_challenge.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=duckduckgo&q=aws+lambda", nil)

	NewHandler(client, tinyserp.DefaultEngines()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}

func TestHandlerNormalizesEngineMapKeys(t *testing.T) {
	html := readFixture(t, "../testdata/bing.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=bing&q=aws+lambda", nil)
	engines := map[string]tinyserp.Engine{
		"  BING  ": tinyserp.BingEngine{},
	}

	NewHandler(client, engines).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}

func TestHandlerReturnsGenericMessageForInternalErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=broken&q=aws+lambda", nil)
	engines := map[string]tinyserp.Engine{
		"broken": failingEngine{},
	}

	NewHandler(nil, engines).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if response["error"] != "internal server error" {
		t.Fatalf("unexpected error message: %s", response["error"])
	}
}

func TestStatusForError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "validation errors are bad request",
			err:  tinyserp.ErrUnsupportedEngine,
			want: http.StatusBadRequest,
		},
		{
			name: "blocked upstream is bad gateway",
			err:  tinyserp.ErrUpstreamBlocked,
			want: http.StatusBadGateway,
		},
		{
			name: "unexpected upstream status is bad gateway",
			err:  tinyserp.ErrUpstreamStatus,
			want: http.StatusBadGateway,
		},
		{
			name: "unknown internal errors are internal server error",
			err:  errors.New("boom"),
			want: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := statusForError(test.err); got != test.want {
				t.Fatalf("unexpected status code: got %d want %d", got, test.want)
			}
		})
	}
}

func TestWriteJSONFallsBackToInternalServerErrorOnEncodingFailure(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusOK, map[string]any{"broken": func() {}})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain;") {
		t.Fatalf("unexpected content type: %s", contentType)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "Internal Server Error") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func readFixture(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}

	return string(content)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type failingEngine struct{}

func (failingEngine) Name() string {
	return "broken"
}

func (failingEngine) BuildRequest(context.Context, string) (*http.Request, error) {
	return nil, errors.New("boom")
}

func (failingEngine) Parse(io.Reader) ([]tinyserp.SearchItem, error) {
	return nil, nil
}
