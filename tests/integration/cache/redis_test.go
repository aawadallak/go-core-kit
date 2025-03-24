package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aawadallak/go-core-kit/core/cache"
	"github.com/aawadallak/go-core-kit/plugin/cache/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisProvider(t *testing.T) {
	ctx := context.Background()

	provider, err := redis.NewRedisProvider(ctx,
		redis.WithAddress("localhost:6379"),
	)
	assert.NoError(t, err)
	defer provider.Close(ctx)

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		operation     func(t *testing.T)
		expectedError error
		expectedValue interface{}
		checkResult   bool
	}{
		{
			name: "Set and Get",
			setup: func(t *testing.T) {
				item := cache.Item{
					Key:       "test-key",
					Value:     []byte("test-value"),
					ExpiresIn: time.Minute,
				}
				err := provider.Set(ctx, item)
				assert.NoError(t, err)
			},
			operation: func(t *testing.T) {
				value, err := provider.Get(ctx, "test-key")
				assert.NoError(t, err)
				assert.Equal(t, []byte("test-value"), value)
			},
			expectedError: nil,
		},
		{
			name: "Get non-existing key",
			operation: func(t *testing.T) {
				_, err := provider.Get(ctx, "non-existing-key")
				assert.ErrorIs(t, err, cache.ErrKeyNotFound)
			},
			expectedError: cache.ErrKeyNotFound,
		},
		{
			name: "Delete existing key",
			setup: func(t *testing.T) {
				item := cache.Item{
					Key:       "delete-key",
					Value:     []byte("delete-value"),
					ExpiresIn: time.Minute,
				}
				err := provider.Set(ctx, item)
				assert.NoError(t, err)
			},
			operation: func(t *testing.T) {
				err := provider.Delete(ctx, "delete-key")
				assert.NoError(t, err)
				_, err = provider.Get(ctx, "delete-key")
				assert.ErrorIs(t, err, cache.ErrKeyNotFound)
			},
			expectedError: nil,
		},
		{
			name: "Delete non-existing key",
			operation: func(t *testing.T) {
				err := provider.Delete(ctx, "non-existing-key")
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}
			tt.operation(t)
		})
	}
}

func TestRedisProviderWithCache(t *testing.T) {
	ctx := context.Background()

	provider, err := redis.NewRedisProvider(ctx,
		redis.WithAddress("localhost:6379"),
	)
	assert.NoError(t, err)
	defer provider.Close(ctx)

	c := cache.New(provider,
		cache.WithEncoder(cache.NewEncoderGzipJSON()),
		cache.WithDecoder(cache.NewDecoderGzipJSON()),
	)

	type Sample struct {
		Value string `json:"value"`
	}

	tests := []struct {
		name          string
		item          cache.Item
		decodeTo      *Sample
		expectedValue Sample
	}{
		{
			name: "Set and Get with struct",
			item: cache.Item{
				Key:       "test-key",
				Value:     Sample{Value: "test-value"},
				ExpiresIn: time.Minute,
			},
			decodeTo:      &Sample{},
			expectedValue: Sample{Value: "test-value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Set(ctx, tt.item)
			assert.NoError(t, err)

			err = c.Get(ctx, tt.item.Key, tt.decodeTo)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValue, *tt.decodeTo)
		})
	}
}

func TestRedisProviderWithCacheResolver_Get(t *testing.T) {
	ctx := context.Background()

	provider, err := redis.NewRedisProvider(ctx,
		redis.WithAddress("localhost:6379"),
	)
	assert.NoError(t, err)
	defer provider.Close(ctx)

	c := cache.New(provider,
		cache.WithEncoder(cache.NewEncoderGzipJSON()),
		cache.WithDecoder(cache.NewDecoderGzipJSON()),
	)

	type Sample struct {
		Value string `json:"value"`
	}

	tests := []struct {
		name          string
		setup         func(t *testing.T) *cache.Resolver[Sample]
		expectedValue Sample
		expectError   bool
	}{
		{
			name: "Get uncached value",
			setup: func(t *testing.T) *cache.Resolver[Sample] {
				return cache.NewResolver("test-key-miss",
					cache.WithCache[Sample](c),
					cache.WithExpiration[Sample](time.Minute),
				)
			},
			expectError: true,
		},
		{
			name: "Get cached value",
			setup: func(t *testing.T) *cache.Resolver[Sample] {
				sample := Sample{Value: "test-value"}
				cItem := cache.Item{
					Key:       "test-key-hit",
					Value:     sample,
					ExpiresIn: time.Minute,
				}
				err := c.Set(ctx, cItem)
				assert.NoError(t, err)
				return cache.NewResolver("test-key-hit",
					cache.WithCache[Sample](c),
					cache.WithExpiration[Sample](time.Minute),
				)
			},
			expectedValue: Sample{Value: "test-value"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := tt.setup(t)
			result, err := resolver.Get(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}
		})
	}
}

func TestRedisProviderWithCacheResolver(t *testing.T) {
	ctx := context.Background()

	provider, err := redis.NewRedisProvider(ctx,
		redis.WithAddress("localhost:6379"),
	)
	assert.NoError(t, err)
	defer provider.Close(ctx)

	c := cache.New(provider,
		cache.WithEncoder(cache.NewEncoderGzipJSON()),
		cache.WithDecoder(cache.NewDecoderGzipJSON()),
	)

	type Sample struct {
		Value string `json:"value"`
	}

	tests := []struct {
		name          string
		setup         func(t *testing.T) *cache.Resolver[Sample]
		fallback      func(ctx context.Context) (Sample, error)
		expectedValue Sample
		expectError   bool
	}{
		{
			name: "Get cached value",
			setup: func(t *testing.T) *cache.Resolver[Sample] {
				sample := Sample{Value: "test-value"}
				cItem := cache.Item{
					Key:       "test-key",
					Value:     sample,
					ExpiresIn: time.Minute,
				}
				err := c.Set(ctx, cItem)
				assert.NoError(t, err)

				return cache.NewResolver("test-key",
					cache.WithCache[Sample](c),
					cache.WithExpiration[Sample](time.Minute),
				)
			},
			fallback: func(ctx context.Context) (Sample, error) {
				return Sample{Value: "test-value-2"}, nil
			},
			expectedValue: Sample{Value: "test-value"},
			expectError:   false,
		},
		{
			name: "Get uncached value using fallback",
			setup: func(t *testing.T) *cache.Resolver[Sample] {
				return cache.NewResolver("uncached-key",
					cache.WithCache[Sample](c),
					cache.WithExpiration[Sample](time.Minute),
				)
			},
			fallback: func(ctx context.Context) (Sample, error) {
				return Sample{Value: "fallback-value"}, nil
			},
			expectedValue: Sample{Value: "fallback-value"},
			expectError:   false,
		},
		{
			name: "Fallback error",
			setup: func(t *testing.T) *cache.Resolver[Sample] {
				return cache.NewResolver("error-key",
					cache.WithCache[Sample](c),
					cache.WithExpiration[Sample](time.Minute),
				)
			},
			fallback: func(ctx context.Context) (Sample, error) {
				return Sample{}, errors.New("fallback error")
			},
			expectedValue: Sample{},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := tt.setup(t)
			result, err := resolver.GetOrFetch(ctx, tt.fallback)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}
		})
	}
}
