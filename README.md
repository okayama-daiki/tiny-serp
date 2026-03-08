# tiny-serp

A tiny SERP (Search Engine Results Page) extraction service in Go.

## What is tiny-serp?

tiny-serp is a minimal SERP extraction service. It queries public search engines, returns structured search data (title, URL, snippet). The goal is to provide a lightweight alternative to commercial search APIs.

It is designed for **private** use and **low-traffic** scenarios, allowing you to deploy your own search API without worrying about costs or maintaining heavy infrastructure.

## Quickstart (HTTP server)

```bash
go run ./cmd/tiny-serp
```

Then, send a search request:

```bash
curl 'http://localhost:8080/search?engine=duckduckgo&q=aws+lambda'
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

## Library usage

```go
service := tinyserp.NewService(tinyserp.DuckDuckGoEngine{}, nil)
response, err := service.Search(ctx, "aws lambda")
```

## Deploy to AWS Lambda

Deployment is handled by `.github/workflows/deploy.yml`.

You need to fork this repository and set up GitHub Actions secrets to deploy to your own AWS account.

Required secrets:

- `AWS_REGION`
- `AWS_DEPLOY_ROLE_ARN`
- `LAMBDA_FUNCTION_NAME`
- `LAMBDA_EXECUTION_ROLE_ARN`

The workflow uses GitHub OIDC for authentication.
To set up the OIDC role and permissions, you may refer to [here](https://github.com/aws-actions/aws-lambda-deploy?tab=readme-ov-file#openid-connect-oidc).

`cmd/lambda` uses `aws-lambda-go/lambdaurl`, so Function URL must use `InvokeMode: RESPONSE_STREAM`.

The recommended GitHub environment name is `production`.

## Future Development

The current release intentionally focuses on minimal SERP extraction.
Potential next areas of development are:

- caching (response reuse and upstream load reduction)
- authentication (optional access control for shared deployments)
- rate limiting (basic abuse prevention)
- timelimit (request timeout control for upstream searches)
- proxies (geo/network routing control)
- headless browsers (fallback for JS-heavy result pages)

All contributions are welcome, but the project will prioritize simplicity and low dependencies over feature bloat. Please open an issue or pull request if you have ideas or improvements!

## License

tiny-serp is licensed under the MIT License. See [LICENSE](LICENSE) for details.
