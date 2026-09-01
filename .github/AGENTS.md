# Maltose CI and Release Instructions

These rules extend the repository root `AGENTS.md` for workflows and release scripts.

## CI Changes

- Preserve validation of the root module, every nested module against the current checkout, and every nested module as a standalone published consumer.
- Keep `GOWORK=off`, `go mod tidy -diff`, `-mod=readonly`, race testing, lint, vulnerability scanning, and the 75 percent root coverage gate unless the task explicitly changes the policy.
- Keep required Redis, MySQL, Apollo, and Nacos service setup aligned with the tests that consume them.
- Use strict shell behavior for multi-step scripts and retain cleanup under success and failure.
- Pin third-party actions and tools to deliberate versions. Review their permissions and avoid expanding token scope without a concrete need.

## Release Changes

- Preserve resumable release behavior and validate immutable tag contents.
- Preserve publication order: root module, first-level modules, then `contrib/observability`.
- Validate every module without repository-local `replace` directives before publication.
- Treat tags and GitHub Releases as externally visible immutable artifacts. Resolve exact versions and targets before mutating release state.
- Keep release operations idempotent so a failed run can resume safely.

## Verification

Run shell syntax checks for changed scripts and exercise their read-only validation modes where available. For workflow changes, compare the local command sequence with `.github/workflows/test.yaml` and `.github/workflows/release.yml`; ensure every existing module and service remains represented.
