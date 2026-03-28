package idem

import "context"

type NoopLocker struct{}

func (NoopLocker) TryLock(_ context.Context, _ string) (acquired bool, release func(context.Context) error, err error) {
	return true, func(context.Context) error { return nil }, nil
}
