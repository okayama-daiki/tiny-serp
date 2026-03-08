# 0006 Handler Service Lifecycle

## Context

`httpapi.NewHandler` currently resolves an engine from the request and then
constructs a fresh `tinyserp.Service` for that request:

- engine lookup is handled once per request from `engine=...`
- `NewService(engine, client)` reads the user agent environment variable and
  applies fixed defaults such as `Accept-Language`

An alternative would be to prebuild one `*tinyserp.Service` per engine when the
handler is created and reuse those instances across requests.

## Decisions

1. Keep per-request `Service` construction for now.
    - `NewService` is lightweight and does not perform network I/O.
    - Search latency is dominated by the upstream HTTP request, so this setup
      cost is negligible in the current scope.
    - This keeps the handler wiring simple and avoids introducing another
      handler-owned registry of prebuilt services.

2. Reuse only the long-lived inputs.
    - The handler already reuses the caller-provided `*http.Client`.
    - The handler also reuses `EngineRegistry` created during handler setup.
    - `Service` stays an ephemeral execution wrapper around one engine plus
      shared configuration defaults.

3. Revisit this if initialization stops being trivial.
    - If `NewService` grows more expensive, accumulates more configuration, or
      starts holding reusable state, the handler should prebuild one service per
      engine and reuse it across requests.
    - That would be a local optimization inside `httpapi.NewHandler`, not a
      change to the public library API.

## Consequences

- The current implementation stays simple and easy to reason about.
- Lambda cold starts and warm invocations behave consistently because handler
  behavior does not depend on an additional prebuilt service cache.
- There is a documented point for future optimization if request setup becomes a
  measurable cost.
