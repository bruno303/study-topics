# AGENTS.md

Operational guide for agents working in this repository.

## Project Overview

- Go learning/practice project focused on context propagation, dependency injection, repository pattern, transactions, Kafka, HTTP APIs, CLI execution, and observability.
- Main executable entrypoints:
  - `cmd/api/main.go`
  - `cmd/cli/main.go`
- Core wiring happens in `internal/setup/container.go` and the app bootstrap in `cmd/api/main.go`.

## Stack Summary

- Language: Go.
- Module: `github.com/bruno303/study-topics/go-study`.
- Declared Go version: `go 1.25.4` in `go.mod`.
- Main libraries in use:
  - `pgx/v5` and `pgxpool` for PostgreSQL access.
  - `confluent-kafka-go/v2` for Kafka.
  - OpenTelemetry SDK plus `otelhttp` for tracing.
  - `slog` adapter for logging.
  - `go-envconfig` and `yaml.v3` for configuration loading.
  - `go.uber.org/mock` for generated mocks.

## Architecture

- `cmd/*` bootstraps config, logging, tracing, dependency injection, workers, Kafka consumers, and the HTTP server.
- `internal/application` contains use cases and app-level interfaces.
  - `internal/application/hello` is the main service layer example.
  - `internal/application/repository` defines repository contracts.
  - `internal/application/transaction` defines `UnitOfWork` and transaction orchestration contracts.
- `internal/infra` contains adapters and runtime integrations.
  - HTTP API and middleware live under `internal/infra/api`.
  - Kafka producer/consumer code lives under `internal/infra/kafka`.
  - PostgreSQL and in-memory repository implementations live under `internal/infra/repository`.
  - OpenTelemetry, slog, correlation-id, and shutdown helpers live under `internal/infra/observability` and `internal/infra/utils`.
  - Workers live under `internal/infra/worker` and write to an outbox before publishing.
- `internal/crosscutting` exposes logging and tracing abstractions used across layers.
- `internal/setup` wires concrete infra implementations into services, handlers, Kafka, and workers.
- Data flow is: transport layer -> application service -> transaction manager -> repository implementation -> external systems.

## Important Folders and Files

- `cmd/api/main.go`: API bootstrap, observability setup, consumer/worker startup, graceful shutdown.
- `cmd/cli/main.go`: CLI bootstrap and logging/tracing setup for local execution.
- `internal/config/config.go`: embedded YAML config loading plus environment overrides and driver selection.
- `internal/setup/container.go`: dependency graph composition.
- `internal/application/hello/service.go`: main use-case implementation and transactional behavior example.
- `internal/application/transaction/transaction.go`: transaction/unit-of-work contract.
- `internal/application/repository/*.go`: repository interfaces.
- `internal/infra/api/hello/api.go`: HTTP route registration and middleware chaining.
- `internal/infra/api/middleware/*`: auth, correlation-id, and request logging middleware.
- `internal/infra/repository/*`: pgx and memdb implementations for hello and outbox data.
- `internal/infra/kafka/*`: producer, consumer, consumer middleware, and message handlers.
- `internal/infra/worker/*`: hello producer worker and outbox sender worker.
- `internal/infra/observability/*`: slog adapter, OTEL setup, trace decorator, correlation-id utilities.
- `config/config.yaml`: default runtime config.
- `config/test.yaml`: test config used by DB/integration-style tests.
- `config/embed.go`: embeds the config directory into the binary.
- `docker-compose.yml`: local Postgres/Kafka/Zookeeper/Kafka UI stack.
- `docker-compose.api.yml`: API container wiring for compose-based local runs.
- `docker/postgres/init.sql`: initial Postgres schema.
- `docker/kafka/init.sh`: topic bootstrap script.
- `tests/integration/main.go`: integration test bootstrap helpers.
- `Makefile`: primary task runner for setup, run, test, and codegen.
- `README.MD`: project notes and feature checklist.

## Development Commands

- Initialize local env file: `make init`
- Download deps and tidy: `make download`
- Vendor deps: `make vendor`
- Run CLI: `make run-cli`
- Run API: `make run-api`
- Debug-build API: `make debug-api`
- Run API with live reload: `make run-api-live`
- Start infra stack: `make docker-up-infra`
- Stop infra stack: `make docker-down`
- Start infra plus API compose stack: `make docker-up-app`
- Run full test suite: `make test`
- Watch tests: `make test-watch`
- Regenerate mocks: `make mocks` or `go generate -v ./...`
- Baseline quality checks: `gofmt -w <changed-files>`, `go vet ./...`, `go test ./...`

## Code Conventions

- Keep changes small and aligned with the existing layering.
- Keep application/business logic out of infra code unless a boundary explicitly requires it.
- Use `context.Context` as the first parameter for request-scoped operations.
- Prefer early returns and wrap errors when that adds useful context.
- Use `panic` only for bootstrap or unrecoverable wiring failures, which is the current pattern in `cmd/*` and setup code.
- Keep interfaces minimal and close to their consumers.
- Follow the existing naming style: exported `PascalCase`, unexported `camelCase`, constructors as `NewX`.
- Use standard Go import grouping and let `gofmt` handle formatting.
- Preserve the repo’s existing compile-time interface assertions where they already exist.
- Regenerate mocks when interfaces with `//go:generate` change.

## Observability

- Logging abstraction: `internal/crosscutting/observability/log`.
- Trace abstraction: `internal/crosscutting/observability/trace`.
- Concrete logger: `internal/infra/observability/slog`.
- Concrete tracer setup: `internal/infra/observability/otel`.
- HTTP requests are wrapped with `otelhttp` in `cmd/api/main.go` and route-tagged in `internal/infra/api/hello/api.go`.
- Kafka consumers use tracing/logging middleware in `internal/infra/kafka/middleware`.
- PostgreSQL query tracing is implemented in `internal/infra/database/pgxpool.go`.
- Log enrichment adds trace IDs and correlation IDs when available.
- Tracing helpers commonly used in the codebase: `trace.Trace`, `trace.InjectError`, `trace.InjectAttributes`, and `tracer.EndTrace` via the tracer interface.

## Infrastructure and External Resources

- Local Postgres is defined in `docker-compose.yml` as `postgres:14.11` with default credentials `postgres/postgres` and database `hello`.
- Local Kafka stack uses Zookeeper, Kafka, topic bootstrap script, and Kafka UI.
- Kafka UI is exposed on port `9001` in `docker-compose.yml`.
- The default SQL schema is created by `docker/postgres/init.sql`.
- The API compose file wires `DATABASE_HOST`, `DATABASE_PORT`, `KAFKA_HOST`, `KAFKA_CONSUMER_GO_STUDY_HOST`, and `APPLICATION_MONITORING_TRACE_URL`.
- Config loading is embedded from `config/*` and overridden by environment variables.
- `CONFIG_FILE` selects the embedded YAML file, and tests use `CONFIG_FILE=test.yaml`.
- `.env.example` only contains `APPLICATION_LOG_LEVEL` and `APPLICATION_LOG_FORMAT`.
- The HTTP example file `requests.http` assumes the default auth secret from config and the API running on `localhost:8080`.

## Testing Strategy

- Unit tests are table-driven where useful and use `gomock`-generated mocks for isolation.
- Package-level tests exist for config loading, service behavior, API handlers, Kafka handlers, workers, and setup wiring.
- Database-backed tests exist for repository behavior and rely on a live Postgres instance.
- Integration helpers live in `tests/integration/main.go` and create the `HELLO_DATA` and `OUTBOX_DATA` tables directly for tests.
- Before DB-dependent tests, start local infra with `make docker-up-infra`.
- Primary verification command is `make test`, which runs `go test -v -timeout 30s -count=1 ./...`.
- For focused iteration, run package-specific `go test` commands instead of the full suite.

## Agent Working Rules

- Inspect `Makefile` before inventing new commands.
- Prefer existing repository patterns over new abstractions.
- Preserve the current layering and dependency direction.
- Do not revert or overwrite user changes outside the task scope.
- Do not add secrets to code, docs, logs, or commit messages.
- Keep ASCII unless the file already requires Unicode.
- Run targeted tests for changed packages; expand to broader checks when the change is larger.
- Update or regenerate mocks when interface changes require it.
- If a needed detail is unclear, confirm it instead of guessing.

## Known Gaps or Unknowns

- No CI/workflow files were found in this repository (`.github` is absent), so CI behavior is unknown.
- No migration tool or migrations directory was found; schema setup appears to rely on SQL bootstrap files and test-time table creation.
- The Dockerfile uses `golang:1.22.3`, which does not match the `go 1.25.4` declaration in `go.mod`; needs confirmation.
- `docker-compose.api.yml` points OTEL to `http://zipkin:9411/api/v2/spans`, but the compose file only has Zipkin commented out and the OTEL code uses OTLP gRPC; needs confirmation.
- README checklist still marks metrics, Loki export, DLT, retry, authentication, and some integration-test work as incomplete.
- `config/config.go` contains an `AsyncCommit` env tag with a tab in the tag value (`ASYNC_COMMIT\t`); this should be treated carefully if the field stops behaving as expected.
