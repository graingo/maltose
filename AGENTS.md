# Maltose Repository Instructions

## Project Scope

Maltose is a lightweight Go Web framework built on Gin. This repository contains the root framework module and several independently published nested modules.

- The root module is `github.com/graingo/maltose`.
- `cmd/maltose/` is the independently published CLI module.
- `contrib/config/apollo/`, `contrib/config/nacos/`, `contrib/metric/otlpmetric/`, `contrib/trace/otlptrace/`, and `contrib/observability/` are independently published modules.
- Packages under `container/`, `database/`, `errors/`, `frame/`, `net/`, `os/`, and `util/` belong to the root module.

Before changing files under `cmd/maltose/`, `contrib/`, or `.github/`, read the nearest `AGENTS.md` in that subtree. These files contain rules that apply in addition to this file.

## Development Principles

- Fix the root cause by replacing incorrect logic with the correct implementation. Do not add temporary patches, compensating conditions, or compatibility branches that preserve incorrect behavior.
- Use direct, affirmative, and precise language in code, comments, documentation, errors, and conclusions.
- Keep control flow clear. Use direct logic for simple behavior and introduce strategies, composition, or factories only for stable business rules or extension points.
- Preserve package responsibilities and reuse existing Maltose components before creating parallel abstractions.
- Keep public APIs small and explicit. Treat exported identifiers, configuration keys, default behavior, error semantics, generated output, and documented examples as compatibility contracts.
- When intentionally changing a public contract, update tests, documentation, examples, and migration guidance in the same change.
- Do not edit files marked `DO NOT EDIT`; change their generator or source definition and regenerate them.
- Format Go changes with `gofmt` or `goimports`. Follow existing short lowercase package names such as `mhttp`, `mdb`, and `mlog`.

## Module and Dependency Rules

- Preserve nested modules as independently consumable Go modules.
- Do not add repository-local `replace` directives to committed `go.mod` files.
- Use `GOWORK=off` when validating how an independently published module resolves its dependencies.
- Keep dependency direction acyclic. `contrib/observability` may compose the OTLP metric and trace modules; leaf modules must not depend on the observability aggregator.
- Add production dependencies only when the standard library and existing dependencies do not provide the required capability.
- Keep framework packages free of application-specific business rules.

## Testing and Verification

Choose validation according to the affected scope:

- Root framework: `go test -race -coverprofile=coverage.out ./...`
- Lint: `golangci-lint run --verbose`
- CLI module: `(cd cmd/maltose && GOWORK=off go test -race -mod=readonly ./...)`
- All nested modules against the current checkout: `.github/scripts/test-local-modules.sh`
- Standalone module integrity: run `GOWORK=off go mod tidy -diff` and `GOWORK=off go test -race -mod=readonly ./...` in every affected module.
- Redis integration: `go test -race -tags=integration ./database/mredis ./contrib/cache/redis`

Apollo and Nacos integration tests require their real services. State any unavailable service explicitly instead of reporting those tests as passed. The root CI coverage threshold is 75 percent.

Add or update table-driven tests beside the package under test. Verify observable behavior and failure semantics, not implementation details alone.

## Review Requirements

After development, complete two review passes and fix discovered problems before handing off.

1. Review behavior and contracts: main flows, lifecycle, failure branches, compatibility, concurrency, transactions, retries, data consistency, boundaries, performance, and sensitive data.
2. Review design and maintainability: package ownership, responsibility, naming, control flow, duplication, dead code, temporary branches, existing capability reuse, and abstraction cost.

Tests and static analysis support these reviews; they do not replace them. Report behavior review, design review, validation results, and remaining risks separately.

## Documentation and Commits

- Update `maltose-docs` when a public API, configuration option, CLI behavior, generated structure, or user workflow changes.
- Update `maltose-quickstart` when the recommended application structure or generated starter behavior changes.
- Keep commits focused and use Conventional Commit prefixes such as `feat:`, `fix:`, `refactor:`, and `chore:`.
- The default branch is `master`. Preserve the current branch unless the task explicitly requests a branch change, and do not mix sibling-repository changes into a Maltose commit.
