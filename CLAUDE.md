# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library (`github.com/aawadallak/go-core-kit`) providing core abstractions and pluggable implementations for common infrastructure concerns: logging, caching, messaging, repositories, configuration, idempotency, and background workers. Go 1.24+. Licensed under GPL v3.0. Status: under active development with potential breaking changes.

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

The codebase follows a **core/plugin separation**:

- **`core/`** — Interfaces and abstractions only (no concrete implementations). Each subdirectory defines a domain contract:
  - `logger` — Logger/Provider interfaces with severity levels
  - `repository` — Generic `AbstractRepository[T]` and `AbstractPaginatedRepository[T, E]` CRUD interfaces
  - `broker` — Publisher/Subscriber/Message/Codec interfaces for messaging
  - `cache` — Cache/Provider interfaces with TTL support
  - `worker` — AbstractWorker with lifecycle hooks (OnStarter, OnShutdowner, OnPreExecutor, OnPostExecutor)
  - `idempotent` — Handler[T]/Repository interfaces for idempotency guarantees
  - `conf` — Provider/ValueMap interfaces for configuration

- **`plugin/`** — Concrete implementations of core interfaces:
  - `logger/zapx` — Zap-based logger
  - `cache/redis` — Redis cache via go-redis
  - `broker/sqs`, `broker/sns` — AWS SQS/SNS messaging
  - `repository/gorm/v2/abstractrepo` — GORM repository with PostgreSQL/MySQL/SQLite support, auto-migration, soft delete, UUID generation
  - `idempotent/storage/gorm`, `idempotent/storage/redis` — Idempotency backends
  - `conf/ssm` — AWS SSM Parameter Store config provider
  - `rest/` — HTTP middleware abstractions

## Key Patterns

- **Options pattern** — All constructors use `NewX(opts ...Option)` with functional options
- **Generics** — Repository and idempotency use Go generics (`AbstractRepository[T]`, `Handler[T]`)
- **Context-first** — All public methods take `context.Context` as first parameter
- **Provider interface** — Core interfaces define a `Provider` for the underlying backend, implementations wrap it
- **Factory functions** — `NewProvider()`, `NewLogger()`, etc. for creating implementations
