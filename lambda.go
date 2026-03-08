package tinyserp

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// LambdaHandler adapts the net/http handler to Lambda Function URLs.
type LambdaHandler struct {
	handler http.Handler
}

// NewLambdaHandler creates a Lambda adapter for the tiny-serp HTTP handler.
func NewLambdaHandler(service *Service) *LambdaHandler {
	return &LambdaHandler{handler: NewHandler(service)}
}

// Handle processes Lambda Function URL requests.
func (h *LambdaHandler) Handle(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
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
	h.handler.ServeHTTP(writer, req)

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
