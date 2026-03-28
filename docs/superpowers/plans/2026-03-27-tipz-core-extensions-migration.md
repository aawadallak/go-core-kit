# Tipz-Core Extensions Migration Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate all extensions from `tipz-core/internal/common/extensions/` into `go-core-kit`, replacing old packages where the tipz-core version is the source of truth.

**Architecture:** Copy-and-adapt migration. Extensions from tipz-core's `internal/common/extensions/{core,plugin}` move to `go-core-kit/{core,plugin}`. The `internal/common` package from tipz-core maps to `go-core-kit/pkg/common` (BaseError, errors, FailureMode, Entity, MustValidateDependencies). Pointer helpers go to `pkg/core/ptr`. All import paths change from `github.com/aawadallak/saas-scaffold-backend/internal/common/...` to `github.com/aawadallak/go-core-kit/...`.

**Tech Stack:** Go 1.24, GORM, AWS SDK v2, NATS, RabbitMQ, Redis, bcrypt, JWX, MessagePack, OpenTelemetry, Chi router, go-resty

---

## Import Path Mapping

| tipz-core import | go-core-kit import |
|---|---|
| `github.com/aawadallak/saas-scaffold-backend/internal/common` | `github.com/aawadallak/go-core-kit/pkg/common` |
| `github.com/aawadallak/saas-scaffold-backend/internal/common/packages/ptr` | `github.com/aawadallak/go-core-kit/pkg/core/ptr` |
| `github.com/aawadallak/saas-scaffold-backend/internal/common/extensions/core/X` | `github.com/aawadallak/go-core-kit/core/X` |
| `github.com/aawadallak/saas-scaffold-backend/internal/common/extensions/plugin/X` | `github.com/aawadallak/go-core-kit/plugin/X` |

## Packages to Skip

- `plugin/event/eventgrpc` - depends on protobuf contracts (`internal/common/contracts/event/v1`) specific to tipz-core. Bring in later when contracts are defined in go-core-kit.

## File Structure

### New files to create:
```
pkg/common/errors.go           -- BaseError, HTTPError, GRPCError, toSnakeCase
pkg/common/types.go            -- All typed errors (ErrInternalServer, ErrResourceNotFound, etc.)
pkg/common/failure_mode.go     -- FailureMode, ClassifyFailureMode
pkg/common/entity.go           -- Entity with GORM tags, BeforeCreate UUID
pkg/common/deps.go             -- MustValidateDependencies
pkg/core/ptr/fn.go             -- New[T]
pkg/core/ptr/now.go            -- Now()
pkg/core/ptr/safe.go           -- Safe[T]
core/audit/audit.go            -- Log, Provider, Service interfaces
core/audit/options.go          -- Orchestrator options
core/audit/orchestrator.go     -- Orchestrator implementation
core/audit/provider.go         -- StandardProvider
core/cache/codec_msgpack.go    -- MsgPack encoder/decoder (add to existing package)
core/cipher/cipher.go          -- Cipher interface + ErrInvalidHash
core/event/event.go            -- Record, Metadata, Dispatcher, Publisher
core/featureflag/feature_flag.go -- State, Service, Provider interfaces
core/featureflag/service.go    -- FeatureFlagService implementation
core/identity/context.go       -- Entity, Organization, context helpers
core/idem/codec.go             -- Codec interface, JSONCodec
core/idem/codec_msgpack.go     -- MsgPackCodec
core/idem/key.go               -- BuildOperationKey
core/idem/key_test.go          -- Key tests
core/idem/locker.go            -- NoopLocker
core/idem/manager.go           -- Manager with Handle[T]
core/idem/manager_test.go      -- Manager tests
core/idem/model.go             -- Status, Record, ClaimOptions, HandleRequest/Result
core/idem/store.go             -- Store, Locker interfaces
core/job/job.go                -- Job, Repository, Handler
core/job/orchestrator.go       -- Orchestrator
core/job/orchestrator_test.go  -- Orchestrator tests
core/seal/seal.go              -- SealedMessage, Sealer, error types
core/txm/manager.go            -- Fn, Manager interface
plugin/abstractrepo/abstract.go            -- AbstractRepository[T]
plugin/abstractrepo/abstract_pagination.go -- AbstractPaginatedRepository[T, E]
plugin/abstractrepo/context.go             -- WithTx, FromContext
plugin/abstractrepo/errors.go              -- ErrInvalidType, ErrInvalidTxType
plugin/abstractrepo/factory.go             -- NewAbstractRepository, options, notDeletedScope
plugin/abstractrepo/locking.go             -- RetryWithBackoff, Version
plugin/broker/nats/consumer.go             -- NATS consumer
plugin/broker/nats/publisher.go            -- NATS publisher
plugin/broker/nats/worker.go               -- NATS typed worker
plugin/broker/natsjetstream/publisher.go   -- JetStream publisher
plugin/broker/natsjetstream/worker.go      -- JetStream worker with DLQ
plugin/broker/rmq/consumer.go              -- RabbitMQ consumer
plugin/broker/rmq/publisher.go             -- RabbitMQ publisher
plugin/broker/rmq/worker.go               -- RabbitMQ worker
plugin/cipher/bcrypt/factory.go            -- bcrypt adapter
plugin/event/publisher.go                  -- Event publisher
plugin/event/outbox/model.go               -- Outbox entry model
plugin/event/outbox/publisher.go           -- Outbox publisher
plugin/event/outbox/repository.go          -- Outbox repository
plugin/event/outbox/worker.go              -- Outbox worker
plugin/event/eventbroker/consumer.go       -- JetStream event consumer
plugin/event/eventbroker/dispatcher.go     -- JetStream event dispatcher
plugin/event/eventhttp/dispatcher.go       -- HTTP event dispatcher
plugin/idem/gorm/store.go                  -- GORM store
plugin/idem/gorm/locker.go                 -- GORM advisory locker
plugin/idem/inmem/store.go                 -- In-memory store
plugin/idem/inmem/locker.go                -- In-memory locker
plugin/idem/postgres/store.go              -- PostgreSQL store
plugin/idem/postgres/locker.go             -- PostgreSQL advisory locker
plugin/job/jorm/factory.go                 -- GORM job repository
plugin/otel/global.go                      -- OTEL tracer/meter helpers
plugin/restchi/server.go                   -- Chi REST server
plugin/restchi/options.go                  -- Server options
plugin/seal/service.go                     -- JWT seal service
plugin/seal/provider/gorm/factory.go       -- GORM seal repository
plugin/txm/txmgorm/manager.go             -- GORM transaction manager
plugin/txm/txmgorm/manager_test.go        -- Transaction manager tests
```

### Files to modify:
```
core/repository/entity.go  -- Remove Entity (moved to pkg/common), keep interfaces + pagination
core/repository/abstract.go -- Update AbstractRepository to use pkg/common.Entity
go.mod                      -- Add new dependencies
```

### Files/directories to remove:
```
core/idempotent/            -- Replaced by core/idem
plugin/idempotent/          -- Replaced by plugin/idem
plugin/repository/gorm/v2/  -- Replaced by plugin/abstractrepo
```

---

### Task 1: Foundation - pkg/common package

**Files:**
- Create: `pkg/common/errors.go`
- Create: `pkg/common/types.go`
- Create: `pkg/common/failure_mode.go`
- Create: `pkg/common/entity.go`
- Create: `pkg/common/deps.go`

- [ ] **Step 1: Create `pkg/common/errors.go`**

```go
package common

import (
	"encoding/json"
	"fmt"
	"unicode"
)

type HTTPError interface {
	StatusCode() int
	ToJSON() ([]byte, error)
}

type GRPCError interface {
	StatusCode() string
}

func toSnakeCase(str string) string {
	result := make([]rune, 0, len(str))
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

type BaseError struct {
	Code       string
	Message    string
	Cause      error
	Attributes map[string]any
}

func (b *BaseError) Error() string {
	if b.Cause != nil {
		return fmt.Sprintf("%s: %s", b.Code, b.Cause.Error())
	}
	return b.Code
}

func (b *BaseError) ToJSON() ([]byte, error) {
	output := map[string]any{
		"code":    b.Code,
		"message": b.Message,
	}
	if b.Cause != nil {
		output["cause"] = b.Cause.Error()
	}
	if len(b.Attributes) > 0 {
		output["attributes"] = b.Attributes
	}
	return json.Marshal(output)
}

func (b *BaseError) StatusCode() int {
	return 422
}
```

- [ ] **Step 2: Create `pkg/common/types.go`**

```go
package common

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type TypeErrInternalServerErr struct{ *BaseError }

func (t *TypeErrInternalServerErr) Is(target error) bool {
	_, ok := target.(*TypeErrInternalServerErr)
	return ok
}

func (t *TypeErrInternalServerErr) StatusCode() int {
	return http.StatusInternalServerError
}

var ErrInternalServer = &TypeErrInternalServerErr{}

func NewErrInternalServer(cause error) error {
	return &TypeErrInternalServerErr{
		BaseError: &BaseError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Oops, something went wrong, please try again or contact support.",
			Cause:   errors.New(strings.ReplaceAll(cause.Error(), "\"", "")),
		},
	}
}

type TypeErrBearerNotFound struct{ *BaseError }

var ErrBearerNotFound = &TypeErrBearerNotFound{}

func (t *TypeErrBearerNotFound) Is(target error) bool {
	_, ok := target.(*TypeErrBearerNotFound)
	return ok
}

func (t *TypeErrBearerNotFound) StatusCode() int { return http.StatusUnauthorized }

func NewErrBearerNotFound() error {
	return &TypeErrBearerNotFound{
		BaseError: &BaseError{
			Code:    "INVALID_BEARER_TOKEN",
			Message: "The authorization header does not contain a valid bearer token",
		},
	}
}

type TypeErrNotAuthorized struct{ *BaseError }

var ErrNotAuthorized = &TypeErrNotAuthorized{}

func (t *TypeErrNotAuthorized) Is(target error) bool {
	_, ok := target.(*TypeErrNotAuthorized)
	return ok
}

func (t *TypeErrNotAuthorized) StatusCode() int { return http.StatusUnauthorized }

func NewErrNotAuthorized(cause error) error {
	return &TypeErrNotAuthorized{
		BaseError: &BaseError{Code: "NOT_AUTHORIZED", Message: "Not authorized", Cause: cause},
	}
}

type TypeErrParseRequest struct{ *BaseError }

func (t *TypeErrParseRequest) Is(target error) bool {
	_, ok := target.(*TypeErrParseRequest)
	return ok
}

func (t *TypeErrParseRequest) StatusCode() int { return http.StatusBadRequest }

var ErrParseRequest = &TypeErrParseRequest{}

func NewErrParseRequest(cause error) error {
	return &TypeErrParseRequest{
		BaseError: &BaseError{
			Code:    "INVALID_REQUEST_BODY",
			Message: "Oops, something went wrong while decoding the body. Please validate the Content-Type or body message",
			Cause:   cause,
		},
	}
}

type TypeErrRequiredFields struct{ *BaseError }

func (t *TypeErrRequiredFields) Is(target error) bool {
	_, ok := target.(*TypeErrRequiredFields)
	return ok
}

func (t *TypeErrRequiredFields) StatusCode() int { return http.StatusBadRequest }

var ErrRequiredFields = &TypeErrRequiredFields{}

func NewErrRequiredFields(cause error) error {
	attr := make(map[string]any)
	var fields validator.ValidationErrors
	if ok := errors.As(cause, &fields); ok {
		for _, field := range fields {
			attr[toSnakeCase(field.Field())] = field.Error()
		}
	}
	return &TypeErrRequiredFields{
		BaseError: &BaseError{
			Code: "MISSING_REQUIRED_FIELDS", Message: "Missing required fields", Attributes: attr,
		},
	}
}

type TypeErrRequiredField struct{ *BaseError }

func (t *TypeErrRequiredField) Is(target error) bool {
	_, ok := target.(*TypeErrRequiredField)
	return ok
}

func (t *TypeErrRequiredField) StatusCode() int { return http.StatusBadRequest }

var ErrRequiredField = &TypeErrRequiredField{}

func NewErrRequiredField(field string) error {
	return &TypeErrRequiredField{
		BaseError: &BaseError{
			Code: "MISSING_REQUIRED_FIELD", Message: "Missing required field",
			Attributes: map[string]any{"field": field},
		},
	}
}

type TypeErrResourceNotFound struct{ *BaseError }

func (t *TypeErrResourceNotFound) Is(target error) bool {
	_, ok := target.(*TypeErrResourceNotFound)
	return ok
}

func (t *TypeErrResourceNotFound) StatusCode() int { return http.StatusNotFound }

var ErrResourceNotFound = &TypeErrResourceNotFound{}

func NewErrResourceNotFound(err error) error {
	return &TypeErrResourceNotFound{
		BaseError: &BaseError{Code: "RESOURCE_NOT_FOUND", Message: "Resource not found", Cause: err},
	}
}

type TypeErrRecoverableError struct{ *BaseError }

func (t *TypeErrRecoverableError) Is(target error) bool {
	_, ok := target.(*TypeErrRecoverableError)
	return ok
}

func (t *TypeErrRecoverableError) FailureMode() FailureMode { return FailureModeRecoverable }

var ErrRecoverableError = &TypeErrRecoverableError{}

func NewErrRecoverableError(cause error) error {
	return &TypeErrRecoverableError{
		BaseError: &BaseError{Code: "RECOVERABLE_ERROR", Message: "Recoverable error", Cause: cause},
	}
}

type TypeErrNonRecoverableError struct{ *BaseError }

func (t *TypeErrNonRecoverableError) Is(target error) bool {
	_, ok := target.(*TypeErrNonRecoverableError)
	return ok
}

func (t *TypeErrNonRecoverableError) FailureMode() FailureMode { return FailureModeNonRecoverable }

var ErrNonRecoverableError = &TypeErrNonRecoverableError{}

func NewErrNonRecoverableError(cause error) error {
	return &TypeErrNonRecoverableError{
		BaseError: &BaseError{Code: "NON_RECOVERABLE_ERROR", Message: "Non-recoverable error", Cause: cause},
	}
}

type TypeErrConflict struct{ *BaseError }

func (e *TypeErrConflict) Is(target error) bool {
	_, ok := target.(*TypeErrConflict)
	return ok
}

func (e *TypeErrConflict) StatusCode() int { return http.StatusConflict }

var ErrConflict = &TypeErrConflict{}

func NewErrConflict(cause error) error {
	return &TypeErrConflict{
		BaseError: &BaseError{Code: "CONFLICT", Message: cause.Error(), Cause: cause},
	}
}
```

- [ ] **Step 3: Create `pkg/common/failure_mode.go`**

```go
package common

import "errors"

type FailureMode string

const (
	FailureModeRecoverable    FailureMode = "recoverable"
	FailureModeNonRecoverable FailureMode = "non_recoverable"
	FailureModeDrop           FailureMode = "drop"
	FailureModeUnknown        FailureMode = "unknown"
)

type FailureModeError interface {
	FailureMode() FailureMode
}

func ClassifyFailureMode(err error) FailureMode {
	if err == nil {
		return FailureModeDrop
	}
	var fm FailureModeError
	if errors.As(err, &fm) {
		mode := fm.FailureMode()
		if mode != "" {
			return mode
		}
		return FailureModeUnknown
	}
	switch {
	case errors.Is(err, ErrParseRequest),
		errors.Is(err, ErrRequiredField),
		errors.Is(err, ErrRequiredFields):
		return FailureModeDrop
	case errors.Is(err, ErrInternalServer):
		return FailureModeRecoverable
	default:
		return FailureModeUnknown
	}
}
```

- [ ] **Step 4: Create `pkg/common/entity.go`**

```go
package common

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	ID         uint       `json:"-" gorm:"primaryKey"`
	ExternalID string     `json:"id" gorm:"uniqueIndex;not null"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at,omitzero"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (e Entity) GetID() uint { return e.ID }

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ExternalID == "" {
		e.ExternalID = uuid.New().String()
	}
	return nil
}
```

- [ ] **Step 5: Create `pkg/common/deps.go`**

```go
package common

import (
	"reflect"
	"strings"
)

func isPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr
}

func isNilPointer(v reflect.Value) bool {
	return isPointer(v) && v.IsNil()
}

func MustValidateDependencies[T any](input T) T {
	v := reflect.ValueOf(input)
	var structValue reflect.Value
	if isPointer(v) {
		if isNilPointer(v) {
			panic("MustValidateDependencies received a nil pointer")
		}
		structValue = v.Elem()
	} else {
		structValue = v
	}
	if structValue.Kind() != reflect.Struct {
		panic("MustValidateDependencies expects a struct or a pointer to a struct")
	}
	var missingFields []string
	structType := structValue.Type()
	for i := range structValue.NumField() {
		fieldValue := structValue.Field(i)
		fieldType := structType.Field(i)
		if tag, ok := fieldType.Tag.Lookup("deps"); ok && tag == "non_required" {
			continue
		}
		switch fieldValue.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Chan, reflect.Slice, reflect.Map:
			if fieldValue.IsNil() {
				missingFields = append(missingFields, fieldType.Name)
			}
		}
	}
	if len(missingFields) > 0 {
		panic("missing required dependencies: " + strings.Join(missingFields, ", "))
	}
	return input
}
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./pkg/common/...`

- [ ] **Step 7: Commit**

```bash
git add pkg/common/
git commit -m "feat: add pkg/common with BaseError, typed errors, FailureMode, Entity, and deps"
```

---

### Task 2: Foundation - pkg/core/ptr package

**Files:**
- Create: `pkg/core/ptr/fn.go`
- Create: `pkg/core/ptr/now.go`
- Create: `pkg/core/ptr/safe.go`

- [ ] **Step 1: Create `pkg/core/ptr/fn.go`**

```go
package ptr

func New[T any](value T) *T {
	return &value
}
```

- [ ] **Step 2: Create `pkg/core/ptr/now.go`**

```go
package ptr

import "time"

func Now() *time.Time {
	return New(time.Now())
}
```

- [ ] **Step 3: Create `pkg/core/ptr/safe.go`**

```go
package ptr

func Safe[T any](v *T) T {
	if v != nil {
		return *v
	}
	var zero T
	return zero
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./pkg/core/ptr/...`

- [ ] **Step 5: Commit**

```bash
git add pkg/core/ptr/
git commit -m "feat: add pkg/core/ptr with New, Now, and Safe helpers"
```

---

### Task 3: Update core/repository to use pkg/common.Entity

**Files:**
- Modify: `core/repository/entity.go`
- Modify: `core/repository/abstract.go`

- [ ] **Step 1: Update `core/repository/entity.go`**

Replace the current Entity definition with a type alias to `pkg/common.Entity`:

```go
package repository

import "github.com/aawadallak/go-core-kit/pkg/common"

// Entity is a type alias for common.Entity for backward compatibility.
type Entity = common.Entity
```

This preserves the `repository.Entity` type for existing consumers while the canonical definition lives in `pkg/common`.

- [ ] **Step 2: Update `core/repository/abstract.go`**

The `AbstractRepositoryEntity` interface requires `GetID() uint`. The current interface returns `uint` but the old Entity had `int32` ID. Since `pkg/common.Entity` now uses `uint`, ensure the interface still matches:

```go
type AbstractRepositoryEntity interface {
	GetID() uint
}
```

No change needed — already matches.

- [ ] **Step 3: Verify compilation**

Run: `go build ./core/repository/...`

- [ ] **Step 4: Commit**

```bash
git add core/repository/
git commit -m "refactor: alias core/repository.Entity to pkg/common.Entity"
```

---

### Task 4: New core packages - audit, cipher, event, identity, txm

**Files:**
- Create: `core/audit/audit.go`, `core/audit/options.go`, `core/audit/orchestrator.go`, `core/audit/provider.go`
- Create: `core/cipher/cipher.go`
- Create: `core/event/event.go`
- Create: `core/identity/context.go`
- Create: `core/txm/manager.go`

- [ ] **Step 1: Create `core/audit/` package**

Copy all 4 files from tipz-core directly. These already import `go-core-kit/core/logger` — no changes needed.

- [ ] **Step 2: Create `core/cipher/cipher.go`**

Adapt: replace `common.BaseError` with `common` from `pkg/common`.

```go
package cipher

import (
	"github.com/aawadallak/go-core-kit/pkg/common"
)

type TypeErrInvalidHash struct {
	*common.BaseError
}

func (e *TypeErrInvalidHash) Is(target error) bool {
	_, ok := target.(*TypeErrInvalidHash)
	return ok
}

var ErrInvalidHash = &TypeErrInvalidHash{}

func NewErrInvalidHash() error {
	return &TypeErrInvalidHash{
		BaseError: &common.BaseError{
			Code:    "INVALID_HASH",
			Message: "The hash is invalid",
		},
	}
}

type Cipher interface {
	Encrypt(string) ([]byte, error)
	Verify(hashedValue []byte, value string) error
}
```

- [ ] **Step 3: Create `core/event/event.go`**

Copy directly from tipz-core — no `common` imports, only `google/uuid`.

- [ ] **Step 4: Create `core/identity/context.go`**

Copy directly — no `common` imports.

- [ ] **Step 5: Create `core/txm/manager.go`**

Copy directly — no `common` imports.

```go
package txm

import "context"

type Fn func(ctx context.Context) error

type Manager interface {
	WithinTransaction(ctx context.Context, fn Fn) error
}
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./core/audit/... ./core/cipher/... ./core/event/... ./core/identity/... ./core/txm/...`

- [ ] **Step 7: Commit**

```bash
git add core/audit/ core/cipher/ core/event/ core/identity/ core/txm/
git commit -m "feat: add audit, cipher, event, identity, and txm core packages"
```

---

### Task 5: New core packages - featureflag, cache msgpack codec

**Files:**
- Create: `core/featureflag/feature_flag.go`
- Create: `core/featureflag/service.go`
- Create: `core/cache/codec_msgpack.go`

- [ ] **Step 1: Create `core/featureflag/` package**

Copy both files from tipz-core. `service.go` imports `go-core-kit/core/logger` (no changes). `feature_flag.go` has no external imports.

- [ ] **Step 2: Create `core/cache/codec_msgpack.go`**

```go
package cache

import (
	"github.com/vmihailenco/msgpack/v5"
)

func NewEncoderMsgPack() Encoder {
	return msgpack.Marshal
}

func NewDecoderMsgPack() Decoder {
	return msgpack.Unmarshal
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./core/featureflag/... ./core/cache/...`

- [ ] **Step 4: Commit**

```bash
git add core/featureflag/ core/cache/codec_msgpack.go
git commit -m "feat: add featureflag service and msgpack cache codec"
```

---

### Task 6: New core package - idem (replaces idempotent)

**Files:**
- Create: `core/idem/codec.go`, `core/idem/codec_msgpack.go`, `core/idem/key.go`, `core/idem/key_test.go`
- Create: `core/idem/locker.go`, `core/idem/manager.go`, `core/idem/manager_test.go`
- Create: `core/idem/model.go`, `core/idem/store.go`

- [ ] **Step 1: Copy files without `common` dependency**

These files copy directly (no `common` imports):
- `codec.go`, `codec_msgpack.go`, `key.go`, `key_test.go`, `locker.go`, `model.go`, `store.go`

- [ ] **Step 2: Create `core/idem/manager.go`**

Adapt: replace `common.ClassifyFailureMode` / `common.FailureModeDrop` with `common` from `pkg/common`.

```go
package idem

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/aawadallak/go-core-kit/pkg/common"
)
```

Then use `common.ClassifyFailureMode(cause)` and `common.FailureModeDrop` — same function names as tipz-core.

- [ ] **Step 3: Copy test files**

`key_test.go`: copy directly — only imports `idem` package itself.

`manager_test.go`: adapt imports:
- `common` → `github.com/aawadallak/go-core-kit/pkg/common`
- `inmem` → `github.com/aawadallak/go-core-kit/plugin/idem/inmem` (depends on Task 11)

Note: `manager_test.go` cannot be verified until Task 11 creates `plugin/idem/inmem`.

- [ ] **Step 4: Verify compilation (excluding manager_test.go)**

Run: `go build ./core/idem/...`

- [ ] **Step 5: Commit**

```bash
git add core/idem/
git commit -m "feat: add core/idem package replacing core/idempotent"
```

---

### Task 7: New core packages - job, seal

**Files:**
- Create: `core/job/job.go`, `core/job/orchestrator.go`, `core/job/orchestrator_test.go`
- Create: `core/seal/seal.go`

- [ ] **Step 1: Create `core/job/job.go`**

Adapt: `common.Entity` → `common.Entity` from `pkg/common` (same name, just different import path).

```go
package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aawadallak/go-core-kit/pkg/common"
)

type Job struct {
	common.Entity
	// ... rest unchanged
}
```

- [ ] **Step 2: Create `core/job/orchestrator.go`**

Adapt: `common.ErrResourceNotFound` → `common.ErrResourceNotFound` from `pkg/common` (same name).

```go
import (
	"github.com/aawadallak/go-core-kit/pkg/common"
	"github.com/aawadallak/go-core-kit/core/logger"
)
```

- [ ] **Step 3: Create `core/job/orchestrator_test.go`**

Adapt: same `common` import update.

- [ ] **Step 4: Create `core/seal/seal.go`**

Adapt: `common.Entity` and `common.BaseError` both come from `pkg/common`.

```go
import (
	"github.com/aawadallak/go-core-kit/pkg/common"
	"github.com/aawadallak/go-core-kit/core/repository"
)

type SealedMessage struct {
	common.Entity
	// ...
}

type SealedMessageRepository = repository.AbstractRepository[SealedMessage]

type TypeErrUsedSignature struct {
	*common.BaseError
}
// ...
```

- [ ] **Step 5: Verify compilation**

Run: `go build ./core/job/... ./core/seal/...`

- [ ] **Step 6: Commit**

```bash
git add core/job/ core/seal/
git commit -m "feat: add job orchestrator and seal core packages"
```

---

### Task 8: New plugin - abstractrepo (replaces plugin/repository/gorm/v2)

**Files:**
- Create: `plugin/abstractrepo/abstract.go`, `plugin/abstractrepo/abstract_pagination.go`
- Create: `plugin/abstractrepo/context.go`, `plugin/abstractrepo/errors.go`
- Create: `plugin/abstractrepo/factory.go`, `plugin/abstractrepo/locking.go`

- [ ] **Step 1: Create all abstractrepo files**

`abstract.go`: adapt `common.NewErrResourceNotFound` → `common.NewErrResourceNotFound` from `pkg/common`.

```go
import (
	"github.com/aawadallak/go-core-kit/pkg/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
```

Copy directly without changes:
- `abstract_pagination.go` — imports `go-core-kit/core/repository`
- `context.go` — only imports gorm
- `errors.go` — no external imports
- `factory.go` — only imports gorm
- `locking.go` — imports `gorm.io/plugin/optimisticlock`

- [ ] **Step 2: Verify compilation**

Run: `go build ./plugin/abstractrepo/...`

- [ ] **Step 3: Commit**

```bash
git add plugin/abstractrepo/
git commit -m "feat: add plugin/abstractrepo replacing gorm/v2/abstractrepo"
```

---

### Task 9: New plugins - broker (nats, natsjetstream, rmq)

**Files:**
- Create: `plugin/broker/nats/{consumer,publisher,worker}.go`
- Create: `plugin/broker/natsjetstream/{publisher,worker}.go`
- Create: `plugin/broker/rmq/{consumer,publisher,worker}.go`

- [ ] **Step 1: Copy all broker plugin files**

These files have NO `common` imports. They import `go-core-kit/core/logger` and their respective messaging libraries. Copy directly with no import changes.

- [ ] **Step 2: Verify compilation**

Run: `go build ./plugin/broker/nats/... ./plugin/broker/natsjetstream/... ./plugin/broker/rmq/...`

- [ ] **Step 3: Commit**

```bash
git add plugin/broker/nats/ plugin/broker/natsjetstream/ plugin/broker/rmq/
git commit -m "feat: add NATS, NATS JetStream, and RabbitMQ broker plugins"
```

---

### Task 10: New plugins - cipher/bcrypt, otel, restchi

**Files:**
- Create: `plugin/cipher/bcrypt/factory.go`
- Create: `plugin/otel/global.go`
- Create: `plugin/restchi/server.go`, `plugin/restchi/options.go`

- [ ] **Step 1: Create `plugin/cipher/bcrypt/factory.go`**

Adapt: `cipher` import → `github.com/aawadallak/go-core-kit/core/cipher`.

- [ ] **Step 2: Copy `plugin/otel/global.go`**

No `common` imports — copy directly.

- [ ] **Step 3: Copy `plugin/restchi/` files**

These import `go-core-kit/core/logger` and `go-chi/chi` — copy directly.

- [ ] **Step 4: Verify compilation**

Run: `go build ./plugin/cipher/... ./plugin/otel/... ./plugin/restchi/...`

- [ ] **Step 5: Commit**

```bash
git add plugin/cipher/ plugin/otel/ plugin/restchi/
git commit -m "feat: add bcrypt cipher, otel helpers, and chi REST server plugins"
```

---

### Task 11: New plugins - idem storage (gorm, inmem, postgres)

**Files:**
- Create: `plugin/idem/gorm/{store,locker}.go`
- Create: `plugin/idem/inmem/{store,locker}.go`
- Create: `plugin/idem/postgres/{store,locker}.go`

- [ ] **Step 1: Copy all idem storage files**

Adapt: replace `saas-scaffold-backend/.../core/idem` → `go-core-kit/core/idem`.

- [ ] **Step 2: Verify compilation**

Run: `go build ./plugin/idem/...`

- [ ] **Step 3: Now verify idem manager tests**

Run: `go test ./core/idem/...`

- [ ] **Step 4: Commit**

```bash
git add plugin/idem/
git commit -m "feat: add idem storage plugins (gorm, inmem, postgres)"
```

---

### Task 12: New plugins - event system

**Files:**
- Create: `plugin/event/publisher.go`
- Create: `plugin/event/outbox/{model,publisher,repository,worker}.go`
- Create: `plugin/event/eventbroker/{consumer,dispatcher}.go`
- Create: `plugin/event/eventhttp/dispatcher.go`

- [ ] **Step 1: Create `plugin/event/publisher.go`**

Adapt:
- `common` → `github.com/aawadallak/go-core-kit/pkg/common`
- `cevent` → `github.com/aawadallak/go-core-kit/core/event`

- [ ] **Step 2: Create `plugin/event/outbox/` files**

Adapt:
- `cevent` → `github.com/aawadallak/go-core-kit/core/event`
- `common.Entity` → `common.Entity` from `pkg/common`
- `abstractrepo` → `github.com/aawadallak/go-core-kit/plugin/abstractrepo`

- [ ] **Step 3: Create `plugin/event/eventbroker/` files**

Adapt:
- `brokerjs` → `github.com/aawadallak/go-core-kit/plugin/broker/natsjetstream`
- `event` → `github.com/aawadallak/go-core-kit/core/event`

- [ ] **Step 4: Create `plugin/event/eventhttp/dispatcher.go`**

Adapt:
- `event` → `github.com/aawadallak/go-core-kit/core/event`

- [ ] **Step 5: Verify compilation**

Run: `go build ./plugin/event/...`

- [ ] **Step 6: Commit**

```bash
git add plugin/event/
git commit -m "feat: add event plugins (publisher, outbox, eventbroker, eventhttp)"
```

---

### Task 13: New plugins - job/jorm, seal, txm/txmgorm

**Files:**
- Create: `plugin/job/jorm/factory.go`
- Create: `plugin/seal/service.go`
- Create: `plugin/seal/provider/gorm/factory.go`
- Create: `plugin/txm/txmgorm/manager.go`, `plugin/txm/txmgorm/manager_test.go`

- [ ] **Step 1: Create `plugin/job/jorm/factory.go`**

Adapt:
- `common` → `github.com/aawadallak/go-core-kit/pkg/common`
- `j` → `github.com/aawadallak/go-core-kit/core/job`
- `abstractrepo` → `github.com/aawadallak/go-core-kit/plugin/abstractrepo`
- `ptr` → `github.com/aawadallak/go-core-kit/pkg/core/ptr`

- [ ] **Step 2: Create `plugin/seal/service.go`**

Adapt:
- `common` → `github.com/aawadallak/go-core-kit/pkg/common` (for `NewErrInternalServer`, `ErrResourceNotFound`, `MustValidateDependencies`)
- `seal` → `github.com/aawadallak/go-core-kit/core/seal`

- [ ] **Step 3: Create `plugin/seal/provider/gorm/factory.go`**

Adapt:
- `seal` → `github.com/aawadallak/go-core-kit/core/seal`
- `abstractrepo` → `github.com/aawadallak/go-core-kit/plugin/abstractrepo`

- [ ] **Step 4: Create `plugin/txm/txmgorm/manager.go` and `manager_test.go`**

Adapt:
- `txm` → `github.com/aawadallak/go-core-kit/core/txm`
- `abstractrepo` → `github.com/aawadallak/go-core-kit/plugin/abstractrepo`

- [ ] **Step 5: Verify compilation**

Run: `go build ./plugin/job/... ./plugin/seal/... ./plugin/txm/...`

- [ ] **Step 6: Commit**

```bash
git add plugin/job/ plugin/seal/ plugin/txm/
git commit -m "feat: add job/jorm, seal service, and txm/txmgorm plugins"
```

---

### Task 14: Remove old packages

**Files:**
- Remove: `core/idempotent/`
- Remove: `plugin/idempotent/`
- Remove: `plugin/repository/gorm/v2/`

- [ ] **Step 1: Remove old idempotent packages**

```bash
rm -rf core/idempotent/ plugin/idempotent/
```

- [ ] **Step 2: Remove old GORM repository**

```bash
rm -rf plugin/repository/gorm/
```

If `plugin/repository/` is empty after this, remove it too.

- [ ] **Step 3: Update any internal imports**

Check if any remaining files (tests, examples) import old packages. Update them.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove old idempotent and gorm/v2 abstractrepo packages"
```

---

### Task 15: Update go.mod and verify

- [ ] **Step 1: Add new dependencies**

```bash
go get github.com/vmihailenco/msgpack/v5
go get github.com/nats-io/nats.go
go get github.com/rabbitmq/amqp091-go
go get golang.org/x/crypto
go get github.com/lestrrat-go/jwx
go get github.com/go-chi/chi/v5
go get github.com/go-resty/resty/v2
go get github.com/go-playground/validator/v10
go get gorm.io/plugin/optimisticlock
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/metric
go get go.opentelemetry.io/otel/trace
```

- [ ] **Step 2: Tidy and verify**

```bash
go mod tidy
go build ./...
```

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Fix any compilation errors.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: update go.mod with new dependencies"
```

---

### Task 16: Update CLAUDE.md

- [ ] **Step 1: Update CLAUDE.md**

Add new packages to the architecture section: `pkg/common`, `pkg/ptr`, and all new core/plugin packages.

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with migrated packages"
```
