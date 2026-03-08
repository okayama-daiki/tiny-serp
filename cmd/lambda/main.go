package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/okayama-daiki/tiny-serp/lambdaadapter"
)

func main() {
	lambda.Start(lambdaadapter.New(nil, nil).Handle)
}
