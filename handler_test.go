package tinyserp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerSearchSuccess(t *testing.T) {
	html := readFixture(t, "testdata/bing.html")
	service := NewService(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=bing&q=aws+lambda", nil)

	NewHandler(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", contentType)
	}

	var response SearchResponse
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
	service := NewService(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})})

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

	handler := NewHandler(service)
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
	html := readFixture(t, "testdata/duckduckgo_challenge.html")
	service := NewService(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/search?engine=duckduckgo&q=aws+lambda", nil)

	NewHandler(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}
