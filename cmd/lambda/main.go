package main

import (
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/okayama-daiki/tiny-serp/httpapi"
)

var startLambdaURL = lambdaurl.Start
var newHTTPHandler = httpapi.NewHandler

func main() {
	startLambdaURL(newHTTPHandler(nil, nil))
}
