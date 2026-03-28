package audit

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/aawadallak/go-core-kit/core/logger"
)

var ErrProviderClosed = errors.New("provider is closed")

type StandardProvider struct {
	close atomic.Bool
}

var _ Provider = (*StandardProvider)(nil)

// Close implements Provider.Close
func (s *StandardProvider) Close(_ context.Context) error {
	if !s.close.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	// Perform any cleanup operations here
	// For example, flush any buffered logs or close connections

	return nil
}

// Flush implements Provider.
func (s *StandardProvider) Flush(ctx context.Context, entries ...Log) error {
	if s.close.Load() {
		return ErrProviderClosed
	}

	for i := range entries {
		logger.Of(ctx).DebugS("Audit::Log", logger.WithValue("message", entries[i]))
	}

	return nil
}

// NewStandardProvider creates a new StandardProvider
func NewStandardProvider() *StandardProvider {
	return &StandardProvider{
		close: atomic.Bool{},
	}
}
