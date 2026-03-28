# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library (`github.com/aawadallak/go-core-kit`) providing core abstractions and pluggable implementations for common infrastructure concerns. Go 1.24+. Licensed under GPL v3.0. Status: under active development with potential breaking changes.

## Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test ./path/to/package -run TestName

# Run integration tests (require Docker)
go test ./tests/integration/...

# Local Redis for development
make redis-local    # start Redis on port 6379
make redis-stop     # stop and remove Redis container
```

## Architecture

The codebase follows a **pkg/core/plugin separation**:

- **`common/`** — Shared foundation types used across core and plugin packages:
  - `BaseError`, typed HTTP errors (ErrResourceNotFound, ErrInternalServer, etc.)
  - `FailureMode` and `ClassifyFailureMode` for error classification
  - `Entity` — base GORM entity with auto UUID, timestamps, soft delete
  - `MustValidateDependencies` — struct dependency validation

- **`core/ptr/`** — Pointer utility helpers: `New[T]`, `Now()`, `Safe[T]`

- **`core/`** — Interfaces and abstractions only (no concrete implementations):
  - `logger` — Logger/Provider interfaces with severity levels
  - `repository` — Generic `AbstractRepository[T]` and `AbstractPaginatedRepository[T, E]` CRUD interfaces
  - `broker` — Publisher/Subscriber/Message/Codec interfaces for messaging
  - `cache` — Cache/Provider interfaces with TTL support + MsgPack codec
  - `worker` — AbstractWorker with lifecycle hooks
  - `conf` — Provider/ValueMap interfaces for configuration
  - `audit` — Batching audit log system with Provider/Service/Orchestrator
  - `cipher` — Cipher interface (Encrypt/Verify) for hashing strategies
  - `event` — Event Record, Metadata, Dispatcher, Publisher interfaces
  - `featureflag` — Feature toggle Service/Provider with caching and auto-sync
  - `identity` — Request context identity (Entity, Organization, context helpers)
  - `idem` — Idempotency framework (Manager, Store, Locker, Codec, key building)
  - `job` — Async job queue with Orchestrator, Repository, Handler interfaces
  - `seal` — Data sealing/integrity via signatures (SealedMessage, Sealer)
  - `txm` — Transaction manager interface

- **`plugin/`** — Concrete implementations of core interfaces:
  - `logger/zapx`, `logger/slogx` — Zap and slog logger backends
  - `cache/redis` — Redis cache via go-redis
  - `broker/sns`, `broker/sqs` — AWS SQS/SNS messaging
  - `broker/nats`, `broker/natsjetstream` — NATS and JetStream messaging
  - `broker/rmq` — RabbitMQ messaging
  - `abstractrepo` — GORM generic repository with pagination, soft delete, optimistic locking, transaction context
  - `idem/gorm`, `idem/inmem`, `idem/postgres` — Idempotency storage backends
  - `event/publisher`, `event/outbox`, `event/eventbroker`, `event/eventhttp` — Event dispatch implementations
  - `cipher/bcrypt` — bcrypt hashing adapter
  - `job/jorm` — GORM job queue repository
  - `seal/` — JWT-based seal service + GORM provider
  - `txm/txmgorm` — GORM transaction manager with retry
  - `otel` — OpenTelemetry tracer/meter helpers
  - `restchi` — Chi-based HTTP server with middleware support
  - `conf/ssm` — AWS SSM Parameter Store config provider

## Key Patterns

- **Options pattern** — All constructors use `NewX(opts ...Option)` with functional options
- **Generics** — Repository, idempotency, broker workers use Go generics
- **Context-first** — All public methods take `context.Context` as first parameter
- **Provider interface** — Core interfaces define a `Provider` for the underlying backend
- **Factory functions** — `NewProvider()`, `NewLogger()`, etc. for creating implementations
- **common.Entity** — Base GORM entity embedded by domain models (job, seal, outbox)
- **Transaction context** — `abstractrepo.WithTx`/`FromContext` propagates GORM transactions via context
