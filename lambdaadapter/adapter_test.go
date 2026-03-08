package lambdaadapter

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleSeparatesSetCookieHeaders(t *testing.T) {
	adapter := &Adapter{handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1; Path=/")
		w.Header().Add("Set-Cookie", "b=2; Path=/")
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Cache-Control", "no-store")
		w.WriteHeader(http.StatusCreated)
	})}

	response, err := adapter.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/search",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{Method: http.MethodGet},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status code: %d", response.StatusCode)
	}
	if len(response.Cookies) != 2 {
		t.Fatalf("unexpected cookies length: %d", len(response.Cookies))
	}
	if response.Cookies[0] != "a=1; Path=/" || response.Cookies[1] != "b=2; Path=/" {
		t.Fatalf("unexpected cookies: %#v", response.Cookies)
	}
	if _, ok := response.Headers["Set-Cookie"]; ok {
		t.Fatal("set-cookie header should not be present in headers map")
	}
	if response.Headers["Cache-Control"] != "no-cache,no-store" {
		t.Fatalf("unexpected cache-control header: %s", response.Headers["Cache-Control"])
	}
}
