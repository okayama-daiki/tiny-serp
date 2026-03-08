# 0003 Transport Package Split

## Context

The repository now contains three concerns:

- reusable search library logic
- HTTP transport
- AWS Lambda transport

Keeping all of them in the module root would make the public package harder to
scan as the project grows.

## Decisions

1. Keep the public library at the module root.
   - External users can import `github.com/okayama-daiki/tiny-serp` directly.
   - We intentionally avoid `internal/` because the search logic may be reused
     by other repositories.

2. Move transport-specific code into subpackages.
   - `httpapi/` owns the `net/http` handler.
   - `lambdaadapter/` owns the Lambda Function URL bridge.
   - The root package remains focused on search execution, parsing, and public
     types.

3. Keep `cmd/` entrypoints thin.
   - `cmd/tiny-serp` only wires the HTTP server.
   - `cmd/lambda` only wires the Lambda adapter.
   - This prevents the repository from becoming messy even though library,
     Lambda, and local server code live together.

## Consequences

- The repository can hold library and application entrypoints without mixing
  their responsibilities.
- If the local server or Lambda path grows significantly, they can evolve
  independently without changing the public library API.
