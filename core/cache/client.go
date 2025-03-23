package cache

import (
	"context"
	"fmt"
	"sync"
)

type cache struct {
	provider Provider
	options  *options
	mu       sync.RWMutex
	isClosed bool
}

var _ Cache = (*cache)(nil)

func (c *cache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if c.isClosed {
		return nil, ErrClosed
	}

	if c.options.useMutex {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}

	value, err := c.provider.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("cache: failed to get key %q: %w", key, err)
	}

	return value, nil
}

func (c *cache) Get(ctx context.Context, key string, target any) error {
	if c.isClosed {
		return ErrClosed
	}

	if c.options.decoder == nil {
		return fmt.Errorf("cache: no decoder configured")
	}

	if c.options.useMutex {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}

	value, err := c.provider.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("cache: failed to get key %q: %w", key, err)
	}

	if err = c.options.decoder(value, target); err != nil {
		return fmt.Errorf("cache: failed to decode key %q: %w", key, err)
	}

	return nil
}

func (c *cache) Set(ctx context.Context, item Item) error {
	if c.isClosed {
		return ErrClosed
	}

	if c.options.useMutex {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	if item.ExpiresIn < 0 {
		item.ExpiresIn = DefaultExpiration
	}

	// If the value is already a byte slice, set it directly
	// This is to avoid unnecessary encoding
	if _, ok := item.Value.([]byte); ok {
		return c.provider.Set(ctx, item)
	}

	if c.options.encoder == nil {
		return fmt.Errorf("cache: no encoder configured")
	}

	val, err := c.options.encoder(item.Value)
	if err != nil {
		return err
	}

	item.Value = val
	if err := c.provider.Set(ctx, item); err != nil {
		return err
	}

	return nil
}

func (c *cache) Delete(ctx context.Context, key string) error {
	if c.isClosed {
		return ErrClosed
	}

	if c.options.useMutex {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	if err := c.provider.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

func (c *cache) Close(ctx context.Context) error {
	if c.options.useMutex {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	if c.isClosed {
		return nil
	}

	if err := c.provider.Close(ctx); err != nil {
		return err
	}

	c.isClosed = true
	return nil
}

// New creates a new cache instance with the specified provider and options.
// If no options are provided, the cache instance is created with default options.
func New(provider Provider, opts ...Option) Cache {
	options := newOption(opts...)

	cache := &cache{
		provider: provider,
		options:  options,
	}

	if provider == nil {
		cache.provider = NewInMemoryCache()
	}

	return cache
}
