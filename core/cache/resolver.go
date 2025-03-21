package cache

import (
	"context"
	"fmt"
	"time"
)

// Handler is a function type that resolves a value of type T
type Handler[T any] func(ctx context.Context) (T, error)

// ResolverOption is a function that configures a Resolver
type ResolverOption[T any] func(*Resolver[T])

// WithExpiration sets the cache expiration duration
func WithExpiration[T any](d time.Duration) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.expiresIn = d
	}
}

// WithCache sets a custom cache implementation
func WithCache[T any](c Cache) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.cache = c
	}
}

// WithResolverEncoder sets a custom encoder for the resolver
func WithResolverEncoder[T any](encoder Encoder) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.encoder = encoder
	}
}

// WithResolverDecoder sets a custom decoder for the resolver
func WithResolverDecoder[T any](decoder Decoder) ResolverOption[T] {
	return func(r *Resolver[T]) {
		r.decoder = decoder
	}
}

// WithReturnErrorOnSet sets whether to return an error if it fails to save the value to the cache
func WithReturnErrorOnSet() ResolverOption[any] {
	return func(r *Resolver[any]) {
		r.returnErrOnSet = true
	}
}

// Resolver provides a simple way to resolve and cache values
type Resolver[T any] struct {
	cache          Cache
	key            string
	expiresIn      time.Duration
	encoder        Encoder
	decoder        Decoder
	returnErrOnSet bool
}

// New creates a new resolver with default settings
func New[T any](key string, opts ...ResolverOption[T]) *Resolver[T] {
	r := &Resolver[T]{
		cache:     Instance(),
		key:       key,
		expiresIn: 1 * time.Hour,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Get retrieves a value from cache or resolves it using the provided handler
func (r *Resolver[T]) Get(ctx context.Context, handler Handler[T]) (T, error) {
	var target T

	// Try to get from cache first
	item, err := r.cache.Get(ctx, r.key)
	if err == nil {
		// If the value is already of type T, use it directly
		if value, ok := item.Value.(T); ok {
			return value, nil
		}

		// If it's a byte slice and we have a decoder, try to decode it
		if bytes, ok := item.Value.([]byte); ok && r.decoder != nil {
			if err := r.decoder(bytes, &target); err == nil {
				return target, nil
			}
		}

		// Add a fallback to return an error for missing decoder or invalid cached value
		return target, fmt.Errorf("failed to decode value for key '%s': %w", r.key, ErrInvalidCachedValue)
	}

	// If not in cache or error occurred, resolve using handler
	value, err := handler(ctx)
	if err != nil {
		return value, err
	}

	// Cache the resolved value directly without encoding
	item = Item{
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

// Delete removes the cached value
func (r *Resolver[T]) Delete(ctx context.Context) error {
	return r.cache.Delete(ctx, r.key)
}
