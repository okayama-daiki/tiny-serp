# 0004 URI Handling

## Context

The project constructs and parses URIs in two places:

- the Lambda Function URL entrypoint needs HTTP request translation between
  Lambda events and `net/http` handlers
- the built-in search engine implementations need to parse and normalize
  outbound and extracted result URLs

URI handling is easy to get subtly wrong if it relies on manual string
concatenation or ad-hoc parsing.

## Decisions

1. Use the Go standard library as the default URI implementation.
    - `net/url` is the official URI parsing and formatting package in Go.
    - For the current requirements, it is sufficient and keeps dependencies low.

2. Avoid manual URI assembly where maintained adapters already exist.
    - The Lambda entrypoint uses `aws-lambda-go/lambdaurl` for request/response
      translation instead of custom URI concatenation logic.
    - URL normalization inside the built-in engine implementations should
      continue to rely on `url.Parse` rather than hand-written parsing logic.

3. Re-evaluate third-party URI libraries only if requirements outgrow `net/url`.
    - Examples would be stricter RFC normalization needs, more advanced escaping
      control, template-heavy URI generation, or compatibility problems that are
      awkward to express with the standard library.
    - A third-party package should only be added when it clearly improves
      correctness or maintainability enough to justify the extra dependency.

## Consequences

- Current code stays dependency-minimal and idiomatic for Go.
- URI logic for search engines remains centralized around standard library
  behavior instead of custom string processing.
- URI translation for Lambda Function URL is delegated to
  `aws-lambda-go/lambdaurl`, reducing custom transport glue code.
- The project still leaves room to adopt a well-known external package later if
  concrete URI requirements become more complex.
