# AGENTS.md

Guidance for coding agents working in this repository.

## Repository Snapshot

- Language: Go (`go 1.25.4` in `go.mod`).
- Module: `github.com/bruno303/study-topics/go-study`.
- Main entrypoints:
  - API: `cmd/api/main.go`
  - CLI: `cmd/cli/main.go`
- Infra helpers:
  - Local infra via Docker Compose (`docker-compose.yml`)
  - App compose (`docker-compose.api.yml`)
- Hot reload config: `.air.toml`
- Core task runner: `Makefile`

## Cursor/Copilot Rules Status

Checked paths:

- `.cursor/rules/`
- `.cursorrules`
- `.github/copilot-instructions.md`

Result: none of these rule files exist in this repository right now.

## Build / Run / Test Commands

Use `Makefile` targets when available.

### Setup

- Initialize local env file:
  - `make init`
- Download and tidy modules:
  - `make download`
- Vendor dependencies (if needed for tooling/repro):
  - `make vendor`

### Run

- Run CLI app:
  - `make run-cli`
- Run API app:
  - `make run-api`
- Build debug API binary and execute:
  - `make debug-api`
- Run API with live reload (`air`):
  - `make run-api-live`

### Docker / Local Infra

- Start infra (Postgres, Kafka, etc):
  - `make docker-up-infra`
- Stop infra:
  - `make docker-down`
- Start app compose stack:
  - `make docker-up-app`

### Test (Primary)

- Run all tests:
  - `make test`
  - Equivalent: `go test -timeout 30s ./...`
- Watch mode:
  - `make test-watch`

### Test (Single Test / Focused)

Use these patterns for faster iteration.

- Run one test function in one package:
  - `go test ./internal/application/hello -run '^TestHello$' -count=1`
- Run one test function by partial match:
  - `go test ./internal/infra/repository -run 'Rollback' -count=1`
- Run one package only:
  - `go test ./pkg/utils/array -count=1`
- Verbose output for one package:
  - `go test -v ./internal/infra/worker -count=1`

### Integration Test Notes

- Some tests in `internal/infra/repository/transaction-manager_test.go` rely on Postgres.
- Test setup uses `tests/integration/main.go` and `config/test.yaml`.
- Ensure local infra is up before running DB-dependent tests:
  - `make docker-up-infra`

### Lint / Static Checks

No dedicated linter target is defined in `Makefile`.

Recommended baseline checks before finalizing changes:

- Format:
  - `gofmt -w <changed-files>`
- Vet:
  - `go vet ./...`
- Optional compile check:
  - `go test ./...`

## Code Generation

- Regenerate mocks after changing interfaces with `//go:generate`:
  - `make mocks`
  - Equivalent: `go generate -v ./...`
- Mock generation uses `go tool mockgen` (from `go.uber.org/mock`).

## Architecture and Directory Conventions

Preserve existing layering and dependency direction.

- `cmd/`: application entrypoints and bootstrap wiring.
- `internal/application/`: use case/business logic and app-level interfaces.
- `internal/infra/`: adapters/integration concerns (API, DB, Kafka, OTEL, workers).
- `internal/crosscutting/`: cross-cutting abstractions (logging, tracing).
- `internal/setup/`: DI/container composition.
- `tests/integration/`: test helpers for integration-style scenarios.
- `pkg/`: reusable utility packages.

Guideline: keep domain/application logic independent from infra details whenever feasible.

## Style Guidelines (Repo-Specific)

### Formatting and File Hygiene

- Always keep code `gofmt`-formatted.
- Keep files ASCII unless the file already requires Unicode.
- Prefer small, focused changes over broad rewrites.

### Imports

- Use standard Go import grouping:
  1) standard library
  2) internal module imports
  3) external dependencies
- Let `gofmt` manage spacing/order inside groups.
- Use import aliases only when necessary for clarity or collisions.

### Naming

- Exported identifiers: `PascalCase`.
- Unexported identifiers: `camelCase`.
- Constructors follow `NewX(...)` pattern.
- Prefer descriptive names over abbreviations (`txManager`, `helloRepository`, etc.).
- Test names follow `TestXxx` and should describe behavior.

### Types and Interfaces

- Define interfaces close to where they are consumed (current pattern in `application`).
- Keep interfaces minimal and behavior-oriented.
- Use compile-time interface assertions where already used:
  - `var _ SomeInterface = (*someImpl)(nil)`
- Follow existing repo use of `any` for generic transactional callbacks where required.

### Context Usage

- Pass `context.Context` as the first parameter in request-scoped operations.
- Propagate context through service/repository/infra boundaries.
- Do not store context in structs.

### Error Handling

- Return errors up the call stack; prefer early returns.
- Add context to errors when useful (prefer wrapping with `%w` in new code).
- Use `panic` only in startup/bootstrap or unrecoverable wiring failures (existing pattern in `cmd/*` and setup).
- In HTTP handlers, convert internal errors to proper HTTP responses and stop execution immediately.

### Logging and Tracing

- Use the repository abstractions:
  - `log.Log()` for logs
  - `trace.Trace(...)` / `trace.InjectError(...)` / `trace.InjectAttributes(...)`
- Keep log messages concise and actionable.
- Avoid logging secrets or sensitive config values.

### Testing

- Prefer table-driven tests when validating multiple cases.
- Keep unit tests deterministic and isolated.
- Use generated mocks (`Mock*`) with `gomock` for dependency isolation.
- For integration tests, ensure setup/cleanup mirrors existing helpers.

### Configuration

- Keep config in `config/*.yaml` plus env overrides.
- Use `CONFIG_FILE` for switching config (e.g., `test.yaml`).
- Do not hardcode environment-specific endpoints in business logic.

## Agent Working Agreements

- Check `Makefile` first before inventing commands.
- Prefer minimal diffs and preserve current architecture.
- Update or regenerate mocks when interfaces change.
- Run targeted tests for changed packages; run full test suite before finishing substantial work.
- If adding new tooling conventions (e.g., a linter), document them here and in `Makefile`.
