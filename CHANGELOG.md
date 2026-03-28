# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`pkg/common/request_context.go`** — Transport-agnostic `RequestContext` (RequestID, TraceID, SpanID) with context helpers.
- **`plugin/cache/msgpack`** — MsgPack encoder/decoder codec for `core/cache` (moved from core to plugin).
- **`plugin/event/eventbroker/transport.go`** — `Transport` and `ConsumerTransport` interfaces for decoupled event dispatching.
- **`plugin/broker/natsjetstream/event_transport.go`** — NATS JetStream adapter implementing eventbroker transport interfaces.
- **`tests/e2e/seal/`** — End-to-end tests with PostgreSQL testcontainer for seal/unseal flow.
- **`tests/e2e/eventbroker/`** — End-to-end tests for event dispatching with in-memory transport.

### Changed

- **`core/audit`** — `Log` struct is now transport-agnostic: HTTP-specific fields (`Method`, `Endpoint`, `StatusCode`, `IP`, `Signature`) replaced with generic `Action` (string) and `Metadata` (map[string]any). Added `NewHTTPLog()` convenience constructor. **Breaking.**
- **`plugin/seal`** — All types moved from `core/seal` into `plugin/seal/types.go`. Seal is now a self-contained plugin, not a core abstraction. **Breaking.**
- **`plugin/event/eventbroker`** — Dispatcher and consumer now accept `Transport`/`ConsumerTransport` interfaces instead of hard-coded NATS JetStream dependency. **Breaking.**
- **`plugin/event/publisher.go`** and **`plugin/event/outbox/publisher.go`** — Updated to use `common.RequestContext` instead of removed `ActivityContext`.

### Removed

- **`core/seal`** — Moved to `plugin/seal`. Too app-specific for core (embeds GORM Entity, niche use case). **Breaking.**
- **`core/identity`** — Removed entirely. User profile model tied to multi-tenant SaaS pattern, not a general infrastructure concern. **Breaking.**
- **`core/cache/codec_msgpack.go`** — Moved to `plugin/cache/msgpack`. External dependency doesn't belong in core. **Breaking** (import path changed).
- **`pkg/common/activity.go`** — Removed `ActivityContext` (web-specific). Replaced by `RequestContext`. **Breaking.**

---

## [Initial] - 2026-03-27

### Added (from initial migration)

- **`pkg/common`** — Shared foundation types: `BaseError` with typed HTTP errors (`ErrResourceNotFound`, `ErrInternalServer`, `ErrConflict`, etc.), `FailureMode` classification, `Entity` base model with GORM integration and auto UUID, `MustValidateDependencies` for struct validation.
- **`pkg/core/ptr`** — Generic pointer helpers: `New[T]`, `Now()`, `Safe[T]`.
- **`core/audit`** — Batching audit log system with configurable providers and flush intervals.
- **`core/cipher`** — Cipher interface for hashing/verification strategies.
- **`core/event`** — Standardized event records with metadata, correlation/trace IDs, and Dispatcher/Publisher interfaces.
- **`core/featureflag`** — Feature toggle service with caching and auto-sync from providers.
- **`core/identity`** — Request context identity extraction with organization/role support.
- **`core/idem`** — Full idempotency framework with Manager, Store, Locker, Codec, and key building.
- **`core/job`** — Async job queue with Orchestrator, panic recovery, and graceful shutdown.
- **`core/seal`** — Data sealing/integrity verification via JWT signatures with nonce-based replay protection.
- **`core/txm`** — Transaction manager interface.
- **`core/cache`** — MsgPack encoder/decoder codec.
- **`plugin/abstractrepo`** — Enhanced GORM repository with pagination, soft delete, transaction context, and optimistic locking.
- **`plugin/broker/nats`** — NATS Core pub/sub with typed workers.
- **`plugin/broker/natsjetstream`** — NATS JetStream with durable consumers and dead-letter queue support.
- **`plugin/broker/rmq`** — RabbitMQ integration.
- **`plugin/cipher/bcrypt`** — bcrypt hashing adapter.
- **`plugin/event/publisher`** — Event publisher with OTEL trace propagation.
- **`plugin/event/outbox`** — Transactional outbox pattern for eventual consistency.
- **`plugin/event/eventbroker`** — NATS JetStream event consumer and dispatcher.
- **`plugin/event/eventhttp`** — HTTP webhook dispatcher with retry and circuit breaker.
- **`plugin/idem/gorm`** — GORM-based idempotency store with PostgreSQL advisory locks.
- **`plugin/idem/inmem`** — In-memory idempotency store for testing and single-process use.
- **`plugin/idem/postgres`** — Raw SQL PostgreSQL idempotency store with advisory locks.
- **`plugin/job/jorm`** — GORM-backed job queue repository.
- **`plugin/seal`** — JWT-based seal service with GORM provider.
- **`plugin/txm/txmgorm`** — GORM transaction manager with exponential backoff retry on deadlocks.
- **`plugin/otel`** — OpenTelemetry global tracer/meter helpers.
- **`plugin/restchi`** — Chi-based HTTP server with middleware and graceful shutdown.

### Changed

- **`core/repository.Entity`** now aliases `pkg/common.Entity` with `uint` ID, GORM struct tags, and auto UUID generation via `BeforeCreate` hook (previously `int32` ID without GORM tags).
- **`plugin/abstractrepo`** replaces `plugin/repository/gorm/v2/abstractrepo` with added transaction context support, optimistic locking, preloads, and configurable soft/hard delete.

### Removed

- **`core/idempotent`** — Replaced by `core/idem` with a more complete idempotency framework.
- **`plugin/idempotent`** — Replaced by `plugin/idem/*` with multiple storage backends.
- **`plugin/repository/gorm/v2/abstractrepo`** — Replaced by `plugin/abstractrepo`.
