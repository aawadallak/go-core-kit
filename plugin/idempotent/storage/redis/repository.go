package idemredis

import (
	"context"
	"errors"

	"github.com/aawadallak/go-core-kit/core/idempotent"
	"github.com/redis/go-redis/v9"
)

var (
	ErrRedisClientRequired = errors.New("redis client is required")
	ErrKeyRequired         = errors.New("key is required")
	ErrPayloadRequired     = errors.New("payload is required")
)

// Repository implements the idempotent.Repository interface using Redis
type Repository struct {
	client *redis.Client
}

// NewRepository creates a new Repository instance with the given Redis client
func NewRepository(client *redis.Client) (*Repository, error) {
	if client == nil {
		return nil, ErrRedisClientRequired
	}

	// Note: Redis must be configured with snapshot persistence (RDB)
	// via config file or commands:
	// CONFIG SET save "3600 1 300 100 60 10000"
	// CONFIG SET appendonly no

	return &Repository{
		client: client,
	}, nil
}

// Save stores an idempotent event in Redis
func (r *Repository) Save(ctx context.Context, item idempotent.EventItem) error {
	if item.Key == "" {
		return ErrKeyRequired
	}
	if item.Payload == nil {
		return ErrPayloadRequired
	}
	result := r.client.SetNX(ctx, item.Key, item.Payload, 0) // 0 means no expiration
	if err := result.Err(); err != nil {
		return err
	}
	return nil
}

// Find retrieves an idempotent event from Redis by its key
func (r *Repository) Find(ctx context.Context, key string) (*idempotent.EventItem, error) {
	if key == "" {
		return nil, ErrKeyRequired
	}

	// Get the payload from Redis
	result := r.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		return nil, err
	}

	payload, err := result.Bytes()
	if err != nil {
		return nil, err
	}

	return &idempotent.EventItem{
		Key:     key,
		Payload: payload,
	}, nil
}
