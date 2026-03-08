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
- `lambdaadapter/`: Lambda Function URL adapter package
- `cmd/tiny-serp`: local HTTP server entrypoint
- `cmd/lambda`: Lambda entrypoint

This keeps the reusable search logic importable from other repositories while
keeping transport-specific code out of the root package.

## Library usage

```go
engines := tinyserp.DefaultEngines()
engine := engines["bing"]
if engine == nil {
    // handle error
}

service := tinyserp.NewService(engine, nil)
response, err := service.Search(ctx, "aws lambda")
```

External packages can add custom engines by populating their own `map[string]tinyserp.Engine`.

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
        "engine": "bing",
        "resultsReturned": 2
    },
    "items": [
        {
            "rank": 1,
            "title": "AWS Lambda - Amazon Web Services",
            "link": "https://aws.amazon.com/lambda/",
            "snippet": "Run code without provisioning or managing servers."
        }
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

The recommended GitHub environment name is `production`.

## Design notes

Implementation decisions are recorded under `docs/design/`.
