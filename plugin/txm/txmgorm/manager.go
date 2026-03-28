// Package txmgorm provides txmgorm functionality.
package txmgorm

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/aawadallak/go-core-kit/core/txm"
	"github.com/aawadallak/go-core-kit/plugin/abstractrepo"
	"gorm.io/gorm"
)

type RetryPredicate func(err error) bool

type Option func(*Manager)

type Manager struct {
	db             *gorm.DB
	maxRetries     int
	initialBackoff time.Duration
	maxBackoff     time.Duration
	multiplier     float64
	shouldRetry    RetryPredicate
}

var _ txm.Manager = (*Manager)(nil)

func New(db *gorm.DB, opts ...Option) (*Manager, error) {
	if db == nil {
		return nil, errors.New("txmgorm: db is nil")
	}

	m := &Manager{
		db:             db,
		maxRetries:     5,
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     2 * time.Second,
		multiplier:     2,
		shouldRetry:    defaultRetryPredicate,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	if m.maxRetries < 0 {
		m.maxRetries = 0
	}
	if m.initialBackoff <= 0 {
		m.initialBackoff = 100 * time.Millisecond
	}
	if m.maxBackoff <= 0 {
		m.maxBackoff = m.initialBackoff
	}
	if m.multiplier < 1 {
		m.multiplier = 1
	}
	if m.shouldRetry == nil {
		m.shouldRetry = defaultRetryPredicate
	}

	return m, nil
}

func WithMaxRetries(v int) Option {
	return func(m *Manager) {
		m.maxRetries = v
	}
}

func WithInitialBackoff(v time.Duration) Option {
	return func(m *Manager) {
		m.initialBackoff = v
	}
}

func WithMaxBackoff(v time.Duration) Option {
	return func(m *Manager) {
		m.maxBackoff = v
	}
}

func WithBackoffMultiplier(v float64) Option {
	return func(m *Manager) {
		m.multiplier = v
	}
}

func WithRetryPredicate(v RetryPredicate) Option {
	return func(m *Manager) {
		m.shouldRetry = v
	}
}

func (m *Manager) WithinTransaction(ctx context.Context, fn txm.Fn) error {
	if fn == nil {
		return errors.New("txmgorm: fn is nil")
	}

	attempts := m.maxRetries + 1
	for attempt := range attempts {
		if _, err := abstractrepo.FromContext(ctx); err == nil {
			return fn(ctx)
		}

		err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(abstractrepo.WithTx(ctx, tx))
		})

		if attempt >= m.maxRetries || !m.shouldRetry(err) {
			return err
		}

		if err := sleepWithContext(ctx, m.backoffForAttempt(attempt)); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) backoffForAttempt(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	value := float64(m.initialBackoff) * math.Pow(m.multiplier, float64(attempt))
	if value > float64(m.maxBackoff) {
		value = float64(m.maxBackoff)
	}

	if value < float64(time.Millisecond) {
		value = float64(time.Millisecond)
	}

	return time.Duration(value)
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func defaultRetryPredicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	msg := strings.ToLower(err.Error())
	retryHints := []string{
		"deadlock",
		"serialization",
		"could not serialize",
		"database is locked",
		"lock wait timeout",
		"too many connections",
		"connection reset",
		"connection refused",
		"broken pipe",
		"temporary",
	}
	for _, hint := range retryHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}

	return false
}
