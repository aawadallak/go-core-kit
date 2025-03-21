package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// DefaultExpiration is the default duration for items stored in
	// the cache to expire.
	DefaultExpiration time.Duration = 0

	// ErrKeyNotFound is returned in Cache.Get and Cache.Delete when the
	// provided key could not be found in cache.
	ErrKeyNotFound error = errors.New("key not found in cache")

	// ErrInvalidCachedValue is returned when the cached value cannot be decoded
	// into the expected type.
	ErrInvalidCachedValue error = errors.New("invalid cached value")
)

// Cache represents a cache interface, which is used to store and retrieve items.
type Cache interface {
	// Set adds an item to the cache.
	// The item is associated with the given key and can be retrieved using the same key.
	Set(ctx context.Context, item Item) error

	// Get retrieves an item from the cache.
	// The item is identified by the given key and can be retrieved using the same key.
	Get(ctx context.Context, key string) (Item, error)

	// Delete removes an item from the cache.
	// The item is identified by the given key and can be removed using the same key.
	Delete(ctx context.Context, key string) error

	// Close closes the cache.
	Close(ctx context.Context) error
}

// Provider represents a cache provider interface, which is used to interact with a cache.
type Provider interface {
	// Set adds an item to the cache with the given time-to-live (TTL).
	// The item is identified by the given key and can be retrieved using the same key.
	Set(ctx context.Context, item Item) error

	// Get retrieves an item from the cache.
	// The item is identified by the given key and can be retrieved using the same key.
	// It returns the value of the item as a byte slice.
	Get(ctx context.Context, key string) ([]byte, error)

	// Delete removes an item from the cache.
	// The item is identified by the given key and can be removed using the same key.
	Delete(ctx context.Context, key string) error

	// Close closes the cache.
	Close(ctx context.Context) error
}
