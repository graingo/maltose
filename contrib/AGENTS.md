# Maltose Contrib Instructions

These rules extend the repository root `AGENTS.md` for adapters and observability integrations under `contrib/`.

## Module Boundaries

- Treat every nested `go.mod` as an independently versioned and independently consumable module.
- Do not commit repository-local `replace` directives.
- Keep leaf adapters focused on one external system or protocol.
- Keep `contrib/observability` as the composition layer for trace and metric integrations. Preserve the release order: root module, leaf modules, then observability.
- Avoid leaking provider-specific types through shared Maltose interfaces unless the integration explicitly exposes an escape hatch.
- Preserve configuration defaults, environment-variable behavior, lifecycle, shutdown, retry, and error semantics as public contracts.

## External Services and Telemetry

- Separate unit-testable adapter behavior from real-service integration tests.
- Preserve `context.Context` cancellation and deadlines across provider calls.
- Close clients, exporters, and background workers deterministically.
- Keep telemetry instrumentation bounded and avoid high-cardinality attributes by default.
- Do not log credentials, tokens, connection strings, or configuration payloads containing secrets.

## Verification

For every affected nested module, run:

```bash
GOWORK=off go mod tidy -diff
GOWORK=off go test -race -mod=readonly ./...
```

Then run the repository-level `.github/scripts/test-local-modules.sh` so the module consumes the current local framework APIs without changing committed module files.

Apollo and Nacos integration tests require their respective services. Nacos real-server integration currently runs without `-race` because the pinned SDK has a reconnect-state race; keep race testing enabled for unit-testable adapter code.
