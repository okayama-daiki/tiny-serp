# tiny-serp

A tiny SERP (Search Engine Results Page) extraction service for serverless environments.

## Scope

`tiny-serp` fetches public search engine HTML, parses organic search results, and returns structured JSON.

Current scope:

- Engines: `duckduckgo`, `bing`
- Endpoint: `GET /search?engine={engine}&q={query}`
- Runtime: Go
- Deployment target: AWS Lambda Function URL
- Deployment method: GitHub Actions with `aws-actions/aws-lambda-deploy`

## Package layout

- module root: public library package `tinyserp`
- `httpapi/`: `net/http` handler package
- `cmd/tiny-serp`: local HTTP server entrypoint
- `cmd/lambda`: Lambda Function URL entrypoint (uses `aws-lambda-go/lambdaurl`)

This keeps the reusable search logic importable from other repositories while
keeping transport-specific code out of the root package.

## Library usage

```go
service := tinyserp.NewService(tinyserp.BingEngine{}, nil)
response, err := service.Search(ctx, "aws lambda")
```

External packages can implement `tinyserp.Engine` directly. The
`map[string]tinyserp.Engine` registry returned by `DefaultEngines()` is mainly
for HTTP-layer engine resolution.

Non-goals for the initial version:

- caching
- authentication
- rate limiting
- proxies
- headless browsers

## Local usage

```bash
go test ./...
go run ./cmd/tiny-serp
```

Then call:

```bash
curl 'http://localhost:8080/search?engine=bing&q=aws+lambda'
```

Example response:

```json
{
  "searchInformation": {
    "query": "aws lambda",
    "engine": "duckduckgo",
    "resultsReturned": 11
  },
  "items": [
    {
      "rank": 1,
      "title": "Official site",
      "link": "https://aws.amazon.com/lambda",
      "snippet": "AWS Lambda"
    },
    {
      "rank": 2,
      "title": "Serverless Computing - AWS Lambda - Amazon Web Services",
      "link": "https://aws.amazon.com/lambda/",
      "snippet": "AWS Lambda is a serverless compute service for running code without having to provision or manage servers. You pay only for the compute time you consume."
    },
    ...
  ]
}
```

## GitHub Actions deploy

Deployment is handled by `.github/workflows/deploy.yml`.

Configure these repository or environment variables first:

- `AWS_REGION`
- `AWS_DEPLOY_ROLE_ARN`
- `LAMBDA_FUNCTION_NAME`
- `LAMBDA_EXECUTION_ROLE_ARN`

The workflow:

- runs `go test ./...`
- builds `./cmd/lambda` as a Linux `arm64` bootstrap binary
- deploys the ZIP artifact with `aws-actions/aws-lambda-deploy`

Function URL configuration and public access policy are managed outside this
workflow.

`cmd/lambda` uses `aws-lambda-go/lambdaurl`, so Function URL must use
`InvokeMode: RESPONSE_STREAM`.

The recommended GitHub environment name is `production`.

## Design notes

Implementation decisions are recorded under `docs/design/`.
