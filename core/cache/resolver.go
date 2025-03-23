package cache

import (
	"context"
	"fmt"
	"time"
)

// Handler defines a function type that resolves a value of type T.
// It takes a context and returns the resolved value and an error.
//
// Parameters:
//
//	ctx - Context for controlling cancellation and timeouts
//
// Returns:
//
//	T - The resolved value
//	error - Any error that occurred during resolution
type Handler[T any] func(ctx context.Context) (T, error)

// ResolverOption defines a function type for configuring Resolver instances.
// It takes a pointer to a Resolver and modifies its configuration.
type ResolverOption[T any] func(*Resolver[T])

// WithExpiration creates an option to set the cache expiration duration.
//
// Parameters:
//
//	d - The duration after which cached values expire
//
// Returns:
//
//	ResolverOption[T] - Configuration function for Resolver
func WithExpiration[T any](d time.Duration) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.expiresIn = d
	}
}

// WithCache creates an option to set a custom cache implementation.
//
// Parameters:
//
//	c - The custom Cache implementation to use
//
// Returns:
//
//	ResolverOption[T] - Configuration function for Resolver
func WithCache[T any](c Cache) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.cache = c
	}
}

// WithReturnErrorOnSet creates an option to configure whether to return errors
// when setting cache values fails.
//
// Returns:
//
//	ResolverOption[any] - Configuration function for Resolver
func WithReturnErrorOnSet() ResolverOption[any] {
	return func(r *Resolver[any]) {
		r.returnErrOnSet = true
	}
}

// Resolver provides a mechanism to resolve and cache values of type T.
// It handles both retrieval from cache and fresh resolution of values.
//
// Fields:
//
//	cache - The underlying cache implementation
//	key - The cache key for storing/retrieving values
//	expiresIn - Duration after which cached values expire
//	returnErrOnSet - Whether to return errors when cache setting fails
type Resolver[T any] struct {
	cache          Cache
	key            string
	expiresIn      time.Duration
	returnErrOnSet bool
}

// NewResolver creates a new Resolver instance with the specified key and options.
// Default settings include a 5-minute expiration and the default cache instance.
//
// Parameters:
//
//	key - The cache key to use for storing/retrieving values
//	opts - Variable number of configuration options
//
// Returns:
//
//	*Resolver[T] - Pointer to the configured Resolver instance
func NewResolver[T any](key string, opts ...ResolverOption[T]) *Resolver[T] {
	r := &Resolver[T]{
		cache:     Instance(),
		key:       key,
		expiresIn: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// GetOrFetch retrieves a value from cache or resolves it using the provided fallback.
// If the value is in cache, it returns immediately. Otherwise, it uses the fallback
// to fetch the value, caches it for future use, and returns it.
//
// Parameters:
//
//	ctx - Context for controlling cancellation and timeouts
//	fallback - Function to resolve the value if not in cache
//
// Returns:
//
//	T - The retrieved or resolved value
//	error - Any error that occurred during retrieval/resolution
func (r *Resolver[T]) GetOrFetch(ctx context.Context, fallback Handler[T]) (T, error) {
	var target T

	// Attempt to retrieve from cache first
	err := r.cache.Get(ctx, r.key, &target)
	if err == nil {
		return target, nil
	}

	if fallback == nil {
		return target, fmt.Errorf("cache: no fallback provided")
	}

	// Resolve using fallback if cache miss or error
	value, err := fallback(ctx)
	if err != nil {
		return value, err
	}

	// Store resolved value in cache
	item := Item{
		Key:       r.key,
		Value:     value,
		ExpiresIn: r.expiresIn,
	}

	if err := r.cache.Set(ctx, item); err != nil {
		if r.returnErrOnSet {
			return value, err
		}
	}

	return value, nil
}

// Get retrieves a value strictly from cache without fallback resolution.
// Unlike GetOrFetch, it does not attempt to resolve the value if not found.
//
// Parameters:
//
//	ctx - Context for controlling cancellation and timeouts
//
// Returns:
//
//	T - The retrieved value
//	error - Any error from cache retrieval
func (r *Resolver[T]) Get(ctx context.Context) (T, error) {
	var target T

	err := r.cache.Get(ctx, r.key, &target)
	if err != nil {
		return target, err
	}

	return target, nil
}
