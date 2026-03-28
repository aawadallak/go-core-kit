package txmgorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	return db
}

func TestWithinTransaction_RetryThenSuccess(t *testing.T) {
	db := newTestDB(t)
	manager, err := New(db,
		WithMaxRetries(3),
		WithInitialBackoff(time.Millisecond),
		WithMaxBackoff(time.Millisecond),
	)
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	attempts := 0
	err = manager.WithinTransaction(context.Background(), func(ctx context.Context) error {
		attempts++
		if attempts == 1 {
			return errors.New("deadlock detected")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected transaction error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWithinTransaction_NoRetryForNonRetryableError(t *testing.T) {
	db := newTestDB(t)
	manager, err := New(db,
		WithMaxRetries(5),
		WithInitialBackoff(time.Millisecond),
		WithMaxBackoff(time.Millisecond),
	)
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	expectedErr := errors.New("validation failed")
	attempts := 0
	err = manager.WithinTransaction(context.Background(), func(ctx context.Context) error {
		attempts++
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestWithinTransaction_StopsOnContextCancelDuringBackoff(t *testing.T) {
	db := newTestDB(t)
	manager, err := New(db,
		WithMaxRetries(5),
		WithInitialBackoff(200*time.Millisecond),
		WithMaxBackoff(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	attempts := 0
	err = manager.WithinTransaction(ctx, func(ctx context.Context) error {
		attempts++
		return errors.New("deadlock")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt before cancel, got %d", attempts)
	}
}
