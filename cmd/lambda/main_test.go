package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/lambda"
	tinyserp "github.com/okayama-daiki/tiny-serp"
)

func TestMainWiresLambdaURLStart(t *testing.T) {
	originalStart := startLambdaURL
	originalHandlerFactory := newHTTPHandler
	t.Cleanup(func() {
		startLambdaURL = originalStart
		newHTTPHandler = originalHandlerFactory
	})

	handlerFactoryCalled := false
	startCalled := false

	newHTTPHandler = func(client *http.Client, engines map[string]tinyserp.Engine) http.Handler {
		handlerFactoryCalled = true
		if client != nil {
			t.Fatal("expected nil client")
		}
		if engines != nil {
			t.Fatal("expected nil engines")
		}
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	startLambdaURL = func(handler http.Handler, _ ...lambda.Option) {
		startCalled = true
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
	}

	main()

	if !handlerFactoryCalled {
		t.Fatal("expected handler factory to be called")
	}
	if !startCalled {
		t.Fatal("expected lambdaurl start to be called")
	}
}
