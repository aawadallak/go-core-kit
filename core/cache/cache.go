package cache

import (
	"context"
	"time"
)

var (
	// DefaultExpiration is the default duration for items stored in
	// the cache to expire.
	DefaultExpiration time.Duration = 0
)

// Cache represents a cache interface, which is used to store and retrieve items.
type Cache interface {
	// Set adds an item to the cache.
	// The item is associated with the given key and can be retrieved using the same key.
	Set(ctx context.Context, item Item) error

	// GetRaw retrieves the raw bytes of an item from the cache.
	GetRaw(ctx context.Context, key string) ([]byte, error)
	// Get retrieves an item from the cache, decoding it into the provided target.
	Get(ctx context.Context, key string, target any) error

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
