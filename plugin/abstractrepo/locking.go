package abstractrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/plugin/optimisticlock"
)

type Version = optimisticlock.Version

func RetryWithBackoff(ctx context.Context, fn func(ctx context.Context) error) error {
	const (
		maxAttempts   = 5
		baseBackoffMs = 25
	)

	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(baseBackoffMs*(attempt+1)) * time.Millisecond):
			}
		}

		//nolint:errcheck // tx may be nil when no transaction context; guarded by nil check below
		tx, _ := FromContext(ctx)

		var spName string
		if tx != nil {
			spName = fmt.Sprintf("sp_retry_%d", attempt)
			if err := tx.SavePoint(spName).Error; err != nil {
				return err
			}
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		if tx != nil {
			_ = tx.RollbackTo(spName).Error // best-effort savepoint rollback during retry
		}
	}

	return lastErr
}
