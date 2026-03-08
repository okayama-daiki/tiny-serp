# 0005 Service Bound To Single Engine

## Context

The initial implementation exposed `Search(ctx, engineName, query)` on the root
library package.

That was convenient for the HTTP layer, but it leaked transport input details
into the library API:

- library callers had to pass engine names as strings
- the service owned both "which engine to use" and "execute a search"
- custom engine support would have required string-based registration and
  lookup inside the same type

The project now needs a cleaner library API while still allowing external users
to add custom engines.

## Decisions

1. Bind `Service` to a single `Engine`.
    - `NewService(engine, client)` constructs a search service for one engine.
    - `Search(ctx, query)` only accepts the query text.
    - This removes the stringly-typed `engineName` parameter from the library
      API.

2. Define engines through a small public interface.
    - `Engine` provides `Name()`, `BuildRequest(...)`, and `Parse(...)`.
    - Built-in engines (`DuckDuckGoEngine`, `BingEngine`) implement that
      interface.
    - External packages can implement their own engines without changing
      `tiny-serp` internals.

3. Keep string-based engine resolution at the HTTP boundary.
    - The `/search?engine=...` endpoint still accepts engine names as strings.
    - The HTTP layer resolves those names through `EngineRegistry`.
    - `NewEngineRegistry` normalizes names once and provides `Resolve(...)`.
    - This keeps transport concerns out of `Service` while preserving the
      existing HTTP API shape.

4. Keep engine registration simple.
    - Built-in engines can be instantiated directly as `DuckDuckGoEngine{}` and
      `BingEngine{}` for library use.
    - `DefaultEngines()` is a convenience for HTTP-layer defaults.
    - External users can extend the registry by supplying their own
      `map[string]Engine` input to `NewEngineRegistry`.
    - This keeps the public API obvious while avoiding request-path
      normalization logic inside handlers.

## Consequences

- Library callers now work with `Engine` values instead of raw engine name
  strings.
- The HTTP handler remains extensible by building `EngineRegistry` from a
  custom `map[string]Engine`.
- Built-in engines stay simple and dependency-minimal.
- Future engine-specific request behavior can be added inside individual engine
  implementations without changing `Service.Search`.
