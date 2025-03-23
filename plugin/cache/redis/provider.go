package redis

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/cache"
	"github.com/redis/go-redis/v9"
)

type cacheProvider struct {
	options *options
	client  *redis.Client
}

var _ cache.Provider = (*cacheProvider)(nil)

func (c *cacheProvider) Get(ctx context.Context, key string) ([]byte, error) {
	cmd := c.client.Get(ctx, key)
	if cmd.Err() == redis.Nil {
		return nil, cache.ErrKeyNotFound
	}
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	res, err := cmd.Bytes()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *cacheProvider) Set(ctx context.Context, item cache.Item) error {
	return c.client.Set(ctx, item.Key, item.Value, item.ExpiresIn).Err()
}

func (c *cacheProvider) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Close implements cache.Provider.
func (c *cacheProvider) Close(_ context.Context) error {
	if c.client == nil {
		return nil
	}

	return c.client.Close()
}

// NewRedisProvider returns a new Redis cache provider with the specified options.
//
// This function creates a new cache provider that uses Redis as the backend store.
// The options parameter can be used to configure various aspects of the Redis client,
// such as the Redis server address, password, pool size, and retry attempts.
//
// Example usage:
//
//	redisProvider, err := NewRedisProvider(
//	    WithAddress("localhost:6379"),
//	    WithPassword("mypassword"),
//	    WithPoolSize(10),
//	    WithMaxRetry(3),
//	)
//
// The returned provider object implements the cache.Provider interface and can be used
// to perform cache operations such as Set, Get, and Delete.
func NewRedisProvider(ctx context.Context, opts ...Option) (cache.Provider, error) {
	provider := &cacheProvider{
		options: newOptions(opts...),
	}

	if err := provider.options.validate(); err != nil {
		return nil, err
	}

	rdb, err := newClient(ctx, provider.options.redisOptions)
	if err != nil {
		return nil, err
	}

	provider.client = rdb

	return provider, nil
}
