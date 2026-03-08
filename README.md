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
- creates or updates the Lambda Function URL
- ensures the public `NONE` auth permissions required for Function URLs

The recommended GitHub environment name is `production`.

## Design notes

Implementation decisions are recorded under `docs/design/`.
