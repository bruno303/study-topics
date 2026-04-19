# AGENTS.md

Operational guide for development agents working in this repository.

## Project Overview

- This repository is a Go learning/practice service that combines HTTP API, Kafka producer/consumer flows, transactional persistence, and observability.
- There are two entrypoints:
  - `cmd/api/main.go` starts HTTP server, Kafka consumers, producer worker, config/log/trace bootstrap, and graceful shutdown listeners.
  - `cmd/cli/main.go` runs a small CLI flow for local logging/tracing experiments.
- Runtime dependency wiring is centralized in `internal/setup/container.go`.
- Main domain example is `hello`: HTTP and Kafka paths both call `internal/application/hello/service.go`.

## Stack Summary

- Language: Go (`go.mod` declares `go 1.25.4`).
- Module: `github.com/bruno303/study-topics/go-study`.
- Core libraries (from `go.mod` and code usage):
  - PostgreSQL: `github.com/jackc/pgx/v5` (`pgxpool`, transactions, query tracer hooks).
  - Migrations: `github.com/pressly/goose/v3` with embedded SQL files.
  - Kafka: `github.com/confluentinc/confluent-kafka-go/v2`.
  - Tracing: OpenTelemetry SDK + OTLP gRPC exporter + `otelhttp`.
  - Logging: standard `log/slog` through a custom adapter.
  - Config: embedded YAML + env override via `github.com/sethvargo/go-envconfig`.
  - Testing/mocks: stdlib `testing`, `go.uber.org/mock`, `sqlmock`.
- Dev tooling:
  - Live reload: `air` via `.air.toml`.
  - Mock generation: `go generate` with `mockgen` directives.

## Architecture

- Layering used in code:
  - `internal/application`: business/use-case interfaces and services.
  - `internal/infra`: transport/integration implementations (HTTP, Kafka, DB, observability, workers).
  - `internal/crosscutting/observability`: logging/tracing abstractions shared across layers.
  - `internal/setup`: composition root and dependency graph.
- Key flow patterns:
  - HTTP: `internal/infra/api/hello` -> middleware chain -> `hello.HelloService`.
  - Kafka consume: `internal/infra/kafka/consumer` -> kafka middleware chain -> message handler -> `hello.HelloService`.
  - Kafka produce: `internal/infra/worker/hello-producer` -> `internal/infra/kafka/producer`.
  - Persistence: `hello.HelloService` uses `transaction.TransactionManager` and `UnitOfWork` to access repositories.
- Data backends:
  - `pgxpool` driver path: real PostgreSQL transaction manager + repository.
  - `memdb` driver path: in-memory repository + transaction manager (for fast/local behavior).
- Migration behavior:
  - `internal/setup/container.go` runs migrations before connecting when DB driver is `pgxpool`.
  - Migrations are embedded SQL under `internal/infra/database/migrations` and executed with Goose advisory lock.

## Important Folders and Files

- `cmd/api/main.go`: API process bootstrap, OTel/log setup, consumers/worker startup, graceful shutdown.
- `cmd/cli/main.go`: CLI bootstrap and logging/tracing sample execution.
- `internal/setup/container.go`: dependency injection and runtime composition.
- `internal/config/config.go`: embedded config loading + env overrides + DB driver normalization.
- `config/config.yaml`: default runtime config.
- `config/test.yaml`: test config (same shape, worker disabled).
- `internal/application/hello/service.go`: core use case logic.
- `internal/application/transaction/transaction.go`: unit-of-work/transaction contracts.
- `internal/application/repository/hello.go`: repository contract.
- `internal/infra/api/hello/api.go`: route registration and middleware chain for `/hello`.
- `internal/infra/api/middleware/*`: auth, correlation id, request logging.
- `internal/infra/kafka/*`: producer, consumer group, consumer chain, handlers, trace carrier.
- `internal/infra/repository/*`: pgx and memdb repository/transaction manager implementations.
- `internal/infra/database/migrations.go`: embedded Goose runner with advisory lock.
- `internal/infra/database/migrations/*.sql`: schema migration files.
- `internal/infra/observability/*`: slog adapter, OTel setup, trace decorators, correlation id helpers.
- `internal/infra/worker/hello-producer.go`: periodic Kafka producer worker.
- `tests/integration/main.go`: DB setup/cleanup helpers for tests that need PostgreSQL.
- `docker-compose.yml`: local infra (Postgres, Zookeeper, Kafka, init-kafka, Kafka UI).
- `docker-compose.api.yml`: app container compose overlay and env wiring.
- `docker/kafka/init.sh`: creates Kafka topic `go-study.hello`.
- `Makefile`: canonical local development/test tasks.
- `requests.http`: example authenticated requests for `/hello`.

## Development Commands

Use `Makefile` commands first.

- `make init`: recreate `.env` from `.env.example`.
- `make download`: `go mod download` + `go mod tidy`.
- `make vendor`: vendor dependencies + tidy.
- `make run-api`: run API entrypoint.
- `make run-cli`: run CLI entrypoint.
- `make run-api-live`: run API with `air` live reload.
- `make debug-api`: debug build and run API binary (`./tmp/api`).
- `make docker-up-infra`: start Postgres/Kafka infra.
- `make docker-down`: stop compose services.
- `make docker-up-app`: start infra then run API compose stack.
- `make test`: run full test suite (`go test -v -timeout 30s -count=1 ./...`).
- `make test-watch`: rerun tests on file changes using `nodemon`.
- `make mocks`: regenerate mocks (`go generate -v ./...`).

Useful direct commands when needed:

- `go test ./path/to/pkg -run TestName`: focused test iteration.
- `gofmt -w <files>` and `go vet ./...`: baseline hygiene checks.

## Code Conventions

- Follow existing layering: application logic in `internal/application`; transport/integration in `internal/infra`.
- Keep `context.Context` as first parameter for request-scoped or IO operations.
- Use small interfaces near consumers (existing pattern in `application/*` packages).
- Keep constructor naming style `NewX`; prefer compile-time interface assertions where already used.
- Panic is currently used in bootstrap/wiring paths (`cmd/*`, `setup`); avoid introducing panic in routine business flow.
- Keep formatting/import order gofmt-compatible.
- When editing interface files with `//go:generate`, regenerate mocks.

## Observability

- Abstractions:
  - Logger interface: `internal/crosscutting/observability/log`.
  - Tracer interface: `internal/crosscutting/observability/trace`.
- Concrete implementations:
  - Logger: `internal/infra/observability/slog`.
  - Tracer + OTel setup: `internal/infra/observability/otel`.
- OTel exporter behavior:
  - Uses OTLP gRPC exporter (`otlptracegrpc.WithEndpoint(cfg.Application.Monitoring.TraceUrl)`), insecure transport.
  - Global propagator is TraceContext + Baggage.
- HTTP tracing:
  - Router wrapped by `otelhttp.NewHandler` in `cmd/api/main.go`.
  - Route tags applied by `otelhttp.WithRouteTag` in `internal/infra/api/hello/api.go`.
- Kafka tracing:
  - Producer injects trace context into Kafka headers (`internal/infra/kafka/kafka-trace`).
  - Consumer middleware extracts trace context and continues spans.
- DB tracing:
  - Custom pgx tracer in `internal/infra/database/pgxpool.go` records SQL/args and errors as span attributes/events.
- Log enrichment:
  - Log adapters append `traceId`, `spanId`, and `correlationId` when present.

## Infrastructure and External Resources

- Docker Compose infra (`docker-compose.yml`):
  - Postgres `postgres:14.11`, DB `hello`, user/password `postgres`.
  - Zookeeper + Kafka + topic initializer + Kafka UI (`localhost:9001`).
- API container (`docker-compose.api.yml`): passes DB and Kafka hosts plus trace URL env var.
- Auth behavior:
  - API key-style check in `internal/infra/api/middleware/authentication-middleware.go` expects `Authorization: Bearer <secret>`.
  - Default secret currently stored in `config/config.yaml` and reused by `requests.http`.
- Config model:
  - YAML files are embedded (`config/embed.go`) and loaded by `CONFIG_FILE` selector.
  - Env vars override YAML (`envconfig` with overwrite enabled).
- Migrations:
  - Managed in-app with embedded Goose migrations, not by external migration CLI in Makefile.

## Testing Strategy

- Test command:
  - `make test` runs `go test` on all packages.
- Unit tests:
  - Service, API handler, Kafka handler/middleware paths, worker, setup, config, utility packages.
  - Mocks generated via `mockgen` are used in application tests.
- DB-backed tests:
  - Repository and transaction-manager tests use `tests/integration.SetupTestDB` and need a live Postgres.
  - Helpers run migrations before tests and cleanup by deleting `HELLO_DATA` rows.
- Migration tests:
  - `internal/infra/database/migrations_test.go` uses `sqlmock` to verify lock/unlock and Goose invocation behavior.
- Practical execution order for DB-related work:
  - `make docker-up-infra` first, then run targeted `go test` packages, then broader `make test`.

## Agent Working Rules

- Start by reading `Makefile`, `config/*.yaml`, and touched package tests before changing behavior.
- Preserve architecture direction: avoid moving application rules into infra packages.
- Prefer minimal deltas in the existing pattern (chain middleware, transaction manager, setup container).
- Do not invent commands or dependencies when repository commands/patterns already exist.
- Do not commit secrets or rotate existing sample secrets unless task explicitly requires it.
- Run targeted tests for changed packages; run full suite for cross-cutting changes.
- If changing config keys or env tags, verify both YAML and env override behavior.
- If behavior depends on external services (Postgres/Kafka/OTel collector), state what was/was not validated locally.

## Known Gaps or Unknowns

- CI is unknown: no `.github/workflows` or other CI config was found.
- Go version mismatch needs confirmation:
  - `go.mod` declares `1.25.4`.
  - `Dockerfile` uses `golang:1.22.3` for build and runtime stages.
- Trace endpoint mismatch needs confirmation:
  - Code exports OTLP gRPC to `APPLICATION_MONITORING_TRACE_URL` (default `localhost:4317`).
  - `docker-compose.api.yml` sets `APPLICATION_MONITORING_TRACE_URL=http://zipkin:9411/api/v2/spans`.
  - Zipkin service is commented out in `docker-compose.yml`.
- `internal/application/outbox` directory exists but currently has no implementation files; intent is Unknown.
- `internal/config/config.go` has a suspicious env tag with embedded tab on `AsyncCommit` (`env:"ASYNC_COMMIT\t"`); verify before relying on env override for this field.
