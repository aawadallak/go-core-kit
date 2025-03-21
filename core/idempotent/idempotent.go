package idempotent

import "context"

type EventItem struct {
	Key     string
	Payload []byte
}

type Hook[T any] func(ctx context.Context) (T, error)

type Handler[T any] interface {
	Wrap(ctx context.Context, key string, fn Hook[T]) (T, error)
}

type Repository interface {
	Find(ctx context.Context, key string) (*EventItem, error)
	Save(ctx context.Context, item EventItem) error
}
