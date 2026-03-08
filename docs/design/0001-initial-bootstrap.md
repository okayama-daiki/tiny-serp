# 0001 Initial Bootstrap

Status: Partially superseded by `0002-github-actions-deploy.md` for deployment and `0005-service-bound-to-single-engine.md` for engine composition.

## Context

The repository started effectively empty. The initial goal is a minimal HTTP API that fetches public search engine HTML, parses organic results, and returns structured JSON from AWS Lambda Function URL.

## Decisions

1. Keep the core as a plain `net/http` handler.
    - `/search?engine={engine}&q={query}` is implemented once and reused by both local execution and Lambda.
    - This keeps local tests and Lambda behavior aligned.

2. Use one small `Service` with engine-specific config.
    - Supported engines are hard-coded in a map: `duckduckgo`, `bing`.
    - Each engine only defines its endpoint and parser function.
    - This avoids a larger plugin abstraction before there is evidence we need one.
    - Superseded later by `0005-service-bound-to-single-engine.md`.

3. Parse HTML with fixed fixtures in tests.
    - Tests use `testdata/*.html` fixtures and a stubbed `http.Client`.
    - This keeps the suite deterministic and respects TDD.
    - Real upstream HTML is too unstable to use directly in tests.

4. Treat upstream blocking as a bad gateway condition.
    - DuckDuckGo currently returns bot challenge pages intermittently.
    - The parser detects the challenge marker and returns `ErrUpstreamBlocked`.
    - The HTTP handler maps that to `502` instead of pretending there were zero results.

5. Keep deployment minimal.
    - Lambda uses a thin adapter over the same HTTP handler.
    - No caching, auth, proxies, rate limiting, or headless browsers are introduced.

## Resulting Structure

- `service.go`: outbound fetch + engine parsers
- `httpapi/handler.go`: `/search` HTTP API
- `lambdaadapter/adapter.go`: Lambda Function URL adapter
- `cmd/tiny-serp`: local HTTP server
- `cmd/lambda`: Lambda entrypoint

## Known Limitations

- Search engine HTML is inherently unstable and may require selector maintenance.
- DuckDuckGo can block or challenge requests depending on traffic patterns.
- Bing links are normalized from redirect URLs when possible, but future format changes may require updates.
