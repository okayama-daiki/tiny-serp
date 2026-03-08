package lambdaadapter

import (
	"bytes"
	"context"
	"net/http"
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
func New(service *tinyserp.Service) *Adapter {
	return &Adapter{handler: httpapi.NewHandler(service)}
}

// Handle processes Lambda Function URL requests.
func (a *Adapter) Handle(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	target := request.RawPath
	if target == "" {
		target = "/"
	}
	if request.RawQueryString != "" {
		target += "?" + request.RawQueryString
	}

	body := strings.NewReader(request.Body)
	req, err := http.NewRequestWithContext(ctx, request.RequestContext.HTTP.Method, target, body)
	if err != nil {
		return events.LambdaFunctionURLResponse{}, err
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	writer := &responseBuffer{header: make(http.Header)}
	a.handler.ServeHTTP(writer, req)

	return events.LambdaFunctionURLResponse{
		StatusCode:      writer.statusCodeOrOK(),
		Headers:         flattenHeaders(writer.header),
		Body:            writer.body.String(),
		IsBase64Encoded: false,
	}, nil
}

type responseBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (w *responseBuffer) Header() http.Header {
	return w.header
}

func (w *responseBuffer) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(payload)
}

func (w *responseBuffer) WriteHeader(status int) {
	w.status = status
}

func (w *responseBuffer) statusCodeOrOK() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func flattenHeaders(header http.Header) map[string]string {
	flattened := make(map[string]string, len(header))
	for key, values := range header {
		flattened[key] = strings.Join(values, ",")
	}
	return flattened
}
