// Package inmem provides inmem functionality.
package inmem

import (
	"context"
	"sync"
)

type Locker struct {
	mu    sync.Mutex
	locks map[string]struct{}
}

func NewLocker() *Locker {
	return &Locker{locks: make(map[string]struct{})}
}

func (l *Locker) TryLock(_ context.Context, key string) (acquired bool, release func(context.Context) error, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.locks[key]; exists {
		return false, nil, nil
	}

	l.locks[key] = struct{}{}
	return true, func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.locks, key)
		return nil
	}, nil
}
