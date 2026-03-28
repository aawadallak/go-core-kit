package idem

import "context"

type Store interface {
	Claim(ctx context.Context, key string, opts ClaimOptions) (ClaimResult, error)
	Complete(ctx context.Context, key string, outcome []byte) (Record, error)
	Fail(ctx context.Context, key string, outcome []byte, status Status) (Record, error)
	Get(ctx context.Context, key string) (Record, bool, error)
}

type Locker interface {
	TryLock(ctx context.Context, key string) (locked bool, unlock func(context.Context) error, err error)
}
