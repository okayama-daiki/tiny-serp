package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	tinyserp "github.com/okayama-daiki/tiny-serp"
)

func main() {
	lambda.Start(tinyserp.NewLambdaHandler(tinyserp.NewService(nil)).Handle)
}
