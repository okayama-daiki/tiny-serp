package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	tinyserp "github.com/okayama-daiki/tiny-serp"
	"github.com/okayama-daiki/tiny-serp/lambdaadapter"
)

func main() {
	lambda.Start(lambdaadapter.New(tinyserp.NewService(nil)).Handle)
}
