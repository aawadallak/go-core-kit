# Go Core Kit

[![Go Reference](https://pkg.go.dev/badge/github.com/aawadallak/go-core-kit.svg)](https://pkg.go.dev/github.com/aawadallak/go-core-kit)
[![Go Report Card](https://goreportcard.com/badge/github.com/aawadallak/go-core-kit)](https://goreportcard.com/report/github.com/aawadallak/go-core-kit)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A modular Go toolkit providing production-ready abstractions and pluggable implementations for building backend services. Designed around clean interface boundaries so you can swap implementations without changing application code.

## Installation

```bash
go get github.com/aawadallak/go-core-kit
```

Import only what you need:

```go
import (
    "github.com/aawadallak/go-core-kit/core/logger"
    "github.com/aawadallak/go-core-kit/plugin/logger/zapx"
)
```

## Architecture

The library follows a strict **core/plugin separation**:

- **`core/`** defines interfaces and contracts — zero external dependencies beyond the standard library
- **`plugin/`** provides concrete implementations backed by real infrastructure
- **`pkg/`** contains shared foundation types and utilities

This means your application code depends only on `core/` interfaces. Swap Redis for in-memory caching, SQS for NATS, or Zap for slog — without touching business logic.

## Packages

### Foundation (`pkg/`)

| Package | Description |
|---------|-------------|
| `pkg/common` | Base entity, typed HTTP errors, failure mode classification, request context, dependency validation |
| `pkg/core/ptr` | Generic pointer helpers: `New[T]`, `Now()`, `Safe[T]` |

### Core Interfaces (`core/`)

| Package | Description |
|---------|-------------|
| `core/logger` | Structured logging with severity levels and context propagation |
| `core/cache` | Key-value caching with TTL, codecs (JSON, GZIP), and `Resolver[T]` for cache-or-fetch |
| `core/broker` | Message publishing and subscribing with pluggable codecs |
| `core/repository` | Generic `AbstractRepository[T]` and `AbstractPaginatedRepository[T, E]` |
| `core/conf` | Configuration loading from multiple providers |
| `core/event` | Event records with correlation/trace IDs, metadata, and Dispatcher/Publisher |
| `core/idem` | Idempotency framework with Manager, Store, Locker, and generic `Handle[T]` |
| `core/job` | Async job queue with Orchestrator, panic recovery, and graceful shutdown |
| `core/worker` | Background worker with lifecycle hooks |
| `core/audit` | Transport-agnostic batching audit log system with generic payload |
| `core/cipher` | Hashing and verification interface |
| `core/featureflag` | Feature toggle service with caching and auto-sync |
| `core/txm` | Transaction manager interface |

### Plugin Implementations (`plugin/`)

| Package | Backend | Implements |
|---------|---------|------------|
| `plugin/logger/zapx` | [Zap](https://github.com/uber-go/zap) | `core/logger` |
| `plugin/logger/slogx` | `log/slog` | `core/logger` |
| `plugin/cache/redis` | [go-redis](https://github.com/redis/go-redis) | `core/cache` |
| `plugin/cache/msgpack` | [msgpack](https://github.com/vmihailenco/msgpack) | `core/cache` codec |
| `plugin/broker/sqs` | AWS SQS | `core/broker` |
| `plugin/broker/sns` | AWS SNS | `core/broker` |
| `plugin/broker/nats` | [NATS](https://nats.io) | `core/broker` |
| `plugin/broker/natsjetstream` | NATS JetStream | `core/broker` (with DLQ) |
| `plugin/broker/rmq` | [RabbitMQ](https://www.rabbitmq.com) | `core/broker` |
| `plugin/abstractrepo` | [GORM](https://gorm.io) | `core/repository` |
| `plugin/conf/ssm` | AWS SSM Parameter Store | `core/conf` |
| `plugin/idem/gorm` | GORM/PostgreSQL | `core/idem` |
| `plugin/idem/inmem` | In-memory | `core/idem` |
| `plugin/idem/postgres` | PostgreSQL (raw SQL) | `core/idem` |
| `plugin/event/*` | Outbox, JetStream, HTTP | `core/event` |
| `plugin/job/jorm` | GORM | `core/job` |
| `plugin/cipher/bcrypt` | bcrypt | `core/cipher` |
| `plugin/seal` | JWT (lestrrat-go/jwx) | Data sealing/signatures |
| `plugin/txm/txmgorm` | GORM | `core/txm` |
| `plugin/otel` | OpenTelemetry | Tracer/Meter helpers |
| `plugin/restchi` | [Chi](https://github.com/go-chi/chi) | HTTP server |

## Quick Start

### Logger

```go
package main

import (
    "context"

    "github.com/aawadallak/go-core-kit/core/logger"
    "github.com/aawadallak/go-core-kit/plugin/logger/zapx"
)

func main() {
    // Set up a Zap-backed logger as the global instance
    provider := zapx.NewProvider()
    log := logger.New(logger.WithProvider(provider))
    logger.SetInstance(log)

    ctx := context.Background()
    logger.Of(ctx).Info("service started")
    logger.Of(ctx).InfoS("request processed",
        logger.WithValue("user_id", "abc-123"),
        logger.WithValue("latency_ms", 42),
    )
}
```

### Repository

```go
package main

import (
    "github.com/aawadallak/go-core-kit/pkg/common"
    "github.com/aawadallak/go-core-kit/plugin/abstractrepo"
    "gorm.io/gorm"
)

type User struct {
    common.Entity
    Name  string
    Email string `gorm:"uniqueIndex"`
}

func NewUserRepo(db *gorm.DB) (*abstractrepo.AbstractRepository[User], error) {
    return abstractrepo.NewAbstractRepository[User](db,
        abstractrepo.WithMigrate(),
    )
}

// repo.Save(ctx, &user)
// repo.FindOne(ctx, &User{Email: "foo@bar.com"})
// repo.Tx(ctx, func(ctx context.Context) error { ... })
```

### Idempotency

```go
result, err := idem.Handle(ctx, manager, idem.HandleRequest[Order]{
    Key:   idem.BuildOperationKey(idem.WithAction("create_order"), idem.WithEntityID(orderID)),
    Owner: "order-service",
    TTL:   30 * time.Second,
    Run: func(ctx context.Context) (Order, error) {
        return createOrder(ctx, orderID)
    },
})
// result.Executed == true on first call, false on replay
// result.Value contains the Order
```

### Event Publishing with Outbox

```go
// Publish events transactionally via the outbox pattern
publisher := outbox.NewPublisher(outboxRepo, abstractrepo.WithTx)

err := txManager.WithinTransaction(ctx, func(ctx context.Context) error {
    // Business logic...
    order, _ := orderRepo.Save(ctx, &order)

    // Event is written to outbox in the same transaction
    return publisher.Publish(ctx, &OrderCreatedEvent{OrderID: order.ExternalID})
})
// The outbox worker picks up pending events and dispatches them
```

## Design Principles

- **Interface-first** — All abstractions are Go interfaces. No framework lock-in.
- **Functional options** — Constructors use `NewX(opts ...Option)` for clean, extensible configuration.
- **Context-first** — Every public method accepts `context.Context` for cancellation and propagation.
- **Generics** — Type-safe repositories, idempotency handlers, and broker workers via Go generics.
- **Zero globals by default** — Optional global instances for convenience (`logger.SetInstance`, `cache.SetInstance`), but everything works with explicit dependency injection.

## Requirements

- Go 1.24+

## Contributing

Contributions are welcome! Please open an issue to discuss significant changes before submitting a PR.

## License

This project is licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0).
