# Maltose CLI Instructions

These rules extend the repository root `AGENTS.md` for the independently published `github.com/graingo/maltose/cmd/maltose` module.

## CLI Contracts

- Treat command names, flags, defaults, positional-argument handling, exit behavior, generated paths, generated source, and diagnostic text as user-facing contracts.
- Keep Cobra command validation explicit. Reject unsupported input before filesystem or database work begins.
- Return actionable errors from generators. Do not write partial or syntactically invalid Go output.
- Preserve existing files unless the command explicitly defines append or regeneration behavior.
- Keep `cli/` responsible for command wiring and input validation, `internal/gen/` responsible for Go source generation, and `internal/openapi/` responsible for OpenAPI parsing and generation.

## Generated Source

- Change templates in `internal/gen/template.go` or their generator implementation instead of editing generated fixtures as the primary fix.
- Preserve the distinction between editable output and output marked `DO NOT EDIT`.
- Format and parse generated Go before writing or appending it.
- When a template changes, verify generated package names, imports, receiver types, pointer semantics, zero-value handling, file permissions, and repeat-run behavior.
- Update `maltose-quickstart` when the recommended generated application structure changes.

## Verification

Run from this module:

```bash
GOWORK=off go mod tidy -diff
GOWORK=off go test -race -mod=readonly ./...
```

For generator changes, add a temporary-directory test that inspects or compiles the generated output. For command changes, update command contract tests. Run the repository-level `.github/scripts/test-local-modules.sh` when the CLI consumes a changed root-module API.
