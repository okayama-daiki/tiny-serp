# 0004 URI Handling

## Context

The project constructs and parses URIs in two places:

- the Lambda Function URL adapter needs to reconstruct a request target from
  `RawPath` and `RawQueryString`
- the search service needs to parse and normalize outbound and extracted result
  URLs

URI handling is easy to get subtly wrong if it relies on manual string
concatenation or ad-hoc parsing.

## Decisions

1. Use the Go standard library as the default URI implementation.
   - `net/url` is the official URI parsing and formatting package in Go.
   - For the current requirements, it is sufficient and keeps dependencies low.

2. Avoid manual URI assembly where `net/url` already provides a structured API.
   - The Lambda adapter should build the request target via `url.URL` and
     `RequestURI()`, not by concatenating path and query strings directly.
   - URL normalization inside the service should continue to rely on `url.Parse`
     rather than hand-written parsing logic.

3. Re-evaluate third-party URI libraries only if requirements outgrow `net/url`.
   - Examples would be stricter RFC normalization needs, more advanced escaping
     control, template-heavy URI generation, or compatibility problems that are
     awkward to express with the standard library.
   - A third-party package should only be added when it clearly improves
     correctness or maintainability enough to justify the extra dependency.

## Consequences

- Current code stays dependency-minimal and idiomatic for Go.
- URI logic remains centralized around standard library behavior instead of
  custom string processing.
- The project still leaves room to adopt a well-known external package later if
  concrete URI requirements become more complex.
