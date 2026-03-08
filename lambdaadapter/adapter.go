package lambdaadapter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	tinyserp "github.com/okayama-daiki/tiny-serp"
	"github.com/okayama-daiki/tiny-serp/httpapi"
)

// Adapter bridges Lambda Function URLs to the tiny-serp HTTP handler.
type Adapter struct {
	handler http.Handler
}

// New creates a Lambda adapter backed by the shared HTTP API.
func New(client *http.Client, engines map[string]tinyserp.Engine) *Adapter {
	return &Adapter{handler: httpapi.NewHandler(client, engines)}
}

// Handle processes Lambda Function URL requests.
func (a *Adapter) Handle(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	target := (&url.URL{
		Path:     defaultPath(request.RawPath),
		RawQuery: request.RawQueryString,
	}).RequestURI()

	body := strings.NewReader(request.Body)
	req, err := http.NewRequestWithContext(ctx, request.RequestContext.HTTP.Method, target, body)
	if err != nil {
		return events.LambdaFunctionURLResponse{}, err
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	a.handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return events.LambdaFunctionURLResponse{}, err
	}

	return events.LambdaFunctionURLResponse{
		StatusCode:      resp.StatusCode,
		Headers:         flattenHeaders(resp.Header),
		Body:            string(payload),
		IsBase64Encoded: false,
		Cookies:         extractCookies(resp.Header),
	}, nil
}

func flattenHeaders(header http.Header) map[string]string {
	flattened := make(map[string]string, len(header))
	for key, values := range header {
		if http.CanonicalHeaderKey(key) == "Set-Cookie" {
			continue
		}
		flattened[key] = strings.Join(values, ",")
	}
	return flattened
}

func extractCookies(header http.Header) []string {
	return slices.Clone(header.Values("Set-Cookie"))
}

func defaultPath(rawPath string) string {
	if rawPath == "" {
		return "/"
	}

	return rawPath
}
