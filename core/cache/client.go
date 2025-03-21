package cache

import (
	"context"
	"sync"
)

type cache struct {
	provider Provider
	options  *options
	mu       sync.RWMutex
	isClosed bool
}

var _ Cache = (*cache)(nil)

func (c *cache) Get(ctx context.Context, key string) (Item, error) {
	if c.isClosed {
		return Item{}, ErrClosed
	}

	if c.options.useMutex {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}

	value, err := c.provider.Get(ctx, key)
	if err != nil {
		return Item{}, err
	}

	item := Item{Key: key, Value: value}

	if c.options.decoder != nil {
		if err = c.options.decoder(value, &item.Value); err != nil {
			return item, err
		}
	}

	return item, nil
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

// NewCache creates a new cache instance with the specified provider and options.
// If no options are provided, the cache instance is created with default options.
func NewCache(provider Provider, opts ...Option) Cache {
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
