package idem

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/aawadallak/go-core-kit/common"
)

var ErrAlreadyInProgress = errors.New("idempotency key already in progress")

type Manager struct {
	store  Store
	locker Locker
	codec  Codec
}

func NewManager(store Store, locker Locker, codec Codec) *Manager {
	if codec == nil {
		codec = JSONCodec{}
	}

	return &Manager{
		store:  store,
		locker: locker,
		codec:  codec,
	}
}

func (m *Manager) Claim(ctx context.Context, key string, opts ClaimOptions) (ClaimResult, error) {
	if opts.TTL <= 0 {
		opts.TTL = 30 * time.Second
	}

	if opts.Owner == "" {
		opts.Owner = "idem"
	}

	return m.store.Claim(ctx, key, opts)
}

func (m *Manager) Complete(ctx context.Context, key string, outcome any) (Record, error) {
	payload, err := m.codec.Marshal(outcome)
	if err != nil {
		return Record{}, err
	}

	return m.store.Complete(ctx, key, payload)
}

func (m *Manager) Fail(ctx context.Context, key string, cause error) (Record, error) {
	status := StatusFailed
	if common.ClassifyFailureMode(cause) == common.FailureModeDrop {
		status = StatusDropped
	}

	payload, err := m.codec.Marshal(map[string]string{
		"error": cause.Error(),
	})
	if err != nil {
		return Record{}, err
	}

	return m.store.Fail(ctx, key, payload, status)
}

func (m *Manager) Get(ctx context.Context, key string) (Record, bool, error) {
	return m.store.Get(ctx, key)
}

func Handle[T any](ctx context.Context, m *Manager, req HandleRequest[T]) (HandleResult[T], error) {
	if m == nil {
		return HandleResult[T]{}, errors.New("idempotency manager is required")
	}

	if req.Run == nil {
		return HandleResult[T]{}, errors.New("idempotency run function is required")
	}
	if req.Key == "" {
		return HandleResult[T]{}, errors.New("idempotency key is required")
	}

	locked, unlock, err := m.locker.TryLock(ctx, req.Key)
	if err != nil {
		return HandleResult[T]{}, err
	}

	if !locked {
		return HandleResult[T]{
			Key:      req.Key,
			Status:   StatusProcessing,
			Executed: false,
			Reused:   false,
		}, nil
	}
	defer func() {
		if unlock != nil {
			_ = unlock(ctx) //nolint:errcheck // best-effort lock release after operation complete
		}
	}()

	claim, err := m.Claim(ctx, req.Key, ClaimOptions{
		Owner: req.Owner,
		TTL:   req.TTL,
	})
	if err != nil {
		return HandleResult[T]{}, err
	}

	if !claim.Acquired {
		result := HandleResult[T]{
			Key:      req.Key,
			Status:   claim.Record.Status,
			Executed: false,
			Reused:   true,
		}

		if claim.Record.Status == StatusCompleted && len(claim.Record.Outcome) > 0 {
			var out T
			if err := m.codec.Unmarshal(claim.Record.Outcome, &out); err == nil {
				result.Value = &out
			}
		}

		return result, nil
	}

	value, err := func() (out T, runErr error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				runErr = panicAsError(recovered)
			}
		}()

		return req.Run(ctx)
	}()
	if err != nil {
		rec, ferr := m.Fail(ctx, req.Key, err)
		if ferr != nil {
			return HandleResult[T]{}, errors.Join(err, ferr)
		}

		return HandleResult[T]{
			Key:      req.Key,
			Status:   rec.Status,
			Executed: true,
			Reused:   false,
		}, err
	}

	rec, err := m.Complete(ctx, req.Key, value)
	if err != nil {
		return HandleResult[T]{}, err
	}

	return HandleResult[T]{
		Key:      req.Key,
		Status:   rec.Status,
		Executed: true,
		Reused:   false,
		Value:    &value,
	}, nil
}

func panicAsError(recovered any) error {
	stack := string(debug.Stack())

	switch v := recovered.(type) {
	case error:
		return fmt.Errorf("idempotency run panic: %w\nstack:\n%s", v, stack)
	default:
		return fmt.Errorf("idempotency run panic: %v\nstack:\n%s", v, stack)
	}
}
